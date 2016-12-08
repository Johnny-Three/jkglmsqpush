package user

import (
	"database/sql"
	"errors"
	"fmt"
	"sync"
	"time"

	. "wbproject/jkglmsgpush/src/enviroment"
	. "wbproject/jkglmsgpush/src/nsq"

	_ "github.com/go-sql-driver/mysql"
)

type Users struct {
	Sl []*Userinfo
}

var lock sync.RWMutex

func (u *Users) GetUser(uid int) *Userinfo {

	for _, v := range u.Sl {
		if v.Userid == uid {
			return v
		}
	}
	return nil
}

//传入Users，从DB中构建出来...
func (u *Users) BuildFromDb(wg *sync.WaitGroup, db1 *sql.DB, db2 *sql.DB) error {
	var userid int
	var st, st1 int64

	qs := fmt.Sprintf(`select userid,unix_timestamp(from_unixtime(starttime,'%%Y-%%m-%%d')) as st,	(CASE  WHEN  starttime > unix_timestamp(DATE_SUB(CURDATE(), INTERVAL 30 DAY)) && starttime <unix_timestamp(DATE_SUB(CURDATE(), INTERVAL 0 DAY)) then unix_timestamp(FROM_UNIXTIME(starttime, '%%Y-%%m-%%d')) ELSE unix_timestamp(DATE_SUB(CURDATE(), INTERVAL 30 DAY)) END) as st1
	from wanbu_health_user_walking_recipes where starttime !=0`)

	fmt.Println(qs)

	rows, err := db1.Query(qs)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {

		wg.Add(1)

		err := rows.Scan(&userid, &st, &st1)
		if err != nil {
			return errors.New("数据库错误")
		}
		go func(wg *sync.WaitGroup, lock *sync.RWMutex, userid int, st, st1 int64) {

			defer wg.Done()
			var user Userinfo

			tmp, err := user.New(userid, st, st1)
			if err != nil {
				wg.Done()
				Logger.Criticalf("user:[%d] set init %s", userid, err.Error())
				return
			}
			//todo ..
			/*
				if err := u.SetFinishStatus(tmp, db2); err != nil {
					wg.Done()
					Logger.Criticalf("user:[%d] set init %s", userid, err.Error())
					return
				}
			*/
			lock.Lock()
			u.Sl = append(u.Sl, tmp)
			lock.Unlock()
		}(wg, &lock, userid, st, st1)
	}

	return nil
}

//出现一个0，退出
func CheckFinish(args ...int8) bool {

	v0 := true
	for _, v := range args {
		if v == 0 {
			v0 = false
			break
		} else {
			continue
		}
	}
	return v0
}

//将用户完成情况从DB中查询出来，按照给定的时间日期，将完成情况算完之后，设置到结构体中
//注意查看设置完成之后，是否真的设置上了，引用传递
func (u *Users) SetFinishStatus(ui *Userinfo, db *sql.DB) error {

	var walkingtime int64
	var t1, t2, t3, t4, t5, t6, t7, t8 int8
	qs := fmt.Sprintf("select walkingtime,task1state,task2state,task3state,task4state,task5state,task6state,task7state,task8state from hmp_walking_tasks_000 where uid = %d and walkingtime >= %d  and walkingtime<= %d ", ui.Userid, ui.Chufang.GetStarttime(), ui.Chufang.GetEndtime())

	rows, err := db.Query(qs)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		err := rows.Scan(&walkingtime, &t1, &t2, &t3, &t4, &t5, &t6, &t7, &t8)
		if err != nil {
			return errors.New("SetFinishStatus()_数据库错误")
		}

		var f int8
		if CheckFinish(t1, t2, t3, t4, t5, t6, t7, t8) {
			f = 1
		}
		if err := ui.Chufang.Set(walkingtime, f); err != nil {
			return err
		}
	}
	return nil
}

//Build User，如果没有，Build出来一个，加进这个slice里
func (u *Users) AppendNew(userid int, st, st1 int64) (int, error) {
	var tmp Userinfo
	ptmp, err := tmp.New(userid, st, st1)
	if err != nil {
		return 0, err
	}
	u.Sl = append(u.Sl, ptmp)
	return len(u.Sl), nil
}

//消费NSQ上传消息，更改某个user的处方完成状态
func (u *Users) ModifyUsersChuFangStatus(ch chan UserWalkdaysStruct) {

	for {
		msg := <-ch
		//用户上传消息，需要看看是否有用户在内存里，在内存里再看时间范围是否符合，计算符合天数的完成任务情况
		//不在内存中则继续循环
		user := u.GetUser(msg.Uid)
		if user == nil {
			continue
		}
		mindate := msg.Walkdays[0].WalkDate
		maxdate := msg.Walkdays[len(msg.Walkdays)-1].WalkDate
		yes, s, e := user.CompareDate(mindate, maxdate)
		if yes {
			Logger.Debugf("userid:[%d],starttime:[%d],endtime:[%d]", msg.Uid, s, e)
			//循环设置
			for i := s; i <= e; i += 86400 {
				//解析符合的消息，查看任务完成状态
				for _, v := range msg.Walkdays {
					if v.WalkDate == i {

						if v.Chufangtotal == v.Chufangfinish && v.Chufangtotal != 0 {
							//set finish ..
							user.Chufang.Set(i, 1)
						} else {
							//set no finish ..
							user.Chufang.Set(i, 0)
						}

					}
				}
			}
			Logger.Debug(user.Chufang.ToString())
			Logger.Debug(user.Chufang.Count())
		} else {
			continue
		}

	}
}

//修改某个user的下载处方时间，如果没有此用户，需要在内存中新建这个用户
func (u *Users) ModifyUsersStarttime(ch chan DownloadMsg) {

	for {

		msg := <-ch
		userid := msg.Userid
		starttime := msg.Starttime

		index := u.FindByUserid(userid)
		//无当前用户
		if index == -1 {
			index, _ = u.AppendNew(userid, starttime, starttime)
		}

		//时间无变化，继续等待
		if u.Sl[index].Starttime == starttime {
			continue
		}
		//修改下载时间后，重新构建处方完成结构..
		if err := u.Sl[index].Chufang.Rebuild(starttime); err != nil {
			Logger.Critical(err.Error())
		}
		Logger.Debug("After Modified ...")
		Logger.Debug(u.Sl[index].Chufang.ToString())
		Logger.Debug(u.Sl[index].Chufang.Count())
		u.Sl[index].Starttime = starttime
	}
}

//返回当前Users数量 ..
func (u *Users) Len() int {
	return len(u.Sl)
}

//-1代表未找到，找到返回slice的index
func (u *Users) FindByUserid(userid int) int {

	for i, v := range u.Sl {
		if v.Userid == userid {
			return i
		}
		continue
	}
	return -1
}

//返回当前Users详情 ..
func (u *Users) ToString(userid int) string {

	for _, v := range u.Sl {
		if v.Userid == userid {
			tmp := fmt.Sprintf("userid:[%d],starttime:[%s],finishcount:[%d]", v.Userid, time.Unix(v.Starttime, 0).Format("2006-01-02"), v.Chufang.Count())
			return tmp + v.Chufang.ToString()
		}
	}
	return "没有这个用户"
}
