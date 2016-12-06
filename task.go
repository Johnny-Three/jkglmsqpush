package main

import (
	"fmt"
	"time"
)

var def = 1000

func TaskGundong(users *Users) {

	for _, v := range users.Sl {
		go func(u *Userinfo) {

			//处理每一个用户..
			ct := GetTimestamp(time.Unix(time.Now().Unix(), 0).Format("2006-01-02"))
			if err := u.Chufang.Changeeveryday(ct); err != nil {
				Logger.Critical("userid:[%d],err:%s", u.Userid, err.Error())
			}
		}(v)
	}
}

func TaskWenjuan1(users *Users, c *Config) {

	var u []*Userinfo

	//从用户中筛选出来，需要在固定时间点触发的..
	for _, v := range users.Sl {

		err, days := v.TimeAfterStart()
		if err != nil {
			Logger.Criticalf("user:[%d],got wrong in TimeAfterStart()", v.Userid)
		}

		switch {
		case days == c.Condition1[0]-1:
			u = append(u, v)
		case days == c.Condition1[1]-1:
			u = append(u, v)
		//超76天后，每隔60天去检查一次..
		case days > c.Condition1[2] && (days+1-c.Condition1[2])%c.Condition1[3] == 0:
			u = append(u, v)
		}
	}
	if len(u) > 0 {

		//推送需要推送的人...写pmlist表...
		err := PushGuding(u, c)
		if err != nil {
			Logger.Critical(err)
		}
		err = WriteQuestionnareTask(u, c, 4)
		if err != nil {
			Logger.Critical(err)
		}
	}
}

func TaskWenjuan2(users *Users, c *Config) {

	var u []*Userinfo

	for _, v := range users.Sl {

		err, days := v.TimeAfterStart()
		if err != nil {
			Logger.Criticalf("user:[%d],got wrong in TimeAfterStart()", v.Userid)
		}
		//每隔多少天去检查一下，如果刚巧是30天
		fmt.Printf("userid:[%d],days:[%d]\n", v.Userid, days)
		if days%c.Condition2[2] == 0 {
			if v.Chufang.Count() > c.Condition2[1] || v.Chufang.Count() < c.Condition2[0] {
				u = append(u, v)
			}
		}
	}
	if len(u) > 0 {

		//推送需要推送的人...写pmlist表...
		err := PushChufang(u, c)
		if err != nil {
			Logger.Critical(err)
		}
		err = WriteQuestionnareTask(u, c, 3)
		if err != nil {
			Logger.Critical(err)
		}
	}
}

func WriteQuestionnareTask(u []*Userinfo, c *Config, t int) error {

	sqlStr := fmt.Sprintf("insert into wanbu_health_questionnaire_task (usertype,category,object,times,pushtimes,pushtime,timestamp) values (2,%d,\"\",1,1,unix_timestamp(),unix_timestamp())", t)

	fmt.Println(sqlStr)

	res, err := c.Db1.Exec(sqlStr)

	if err != nil {
		Logger.Critical(err)
		return err
	}

	taskid, err := res.LastInsertId()
	if err != nil {
		Logger.Critical(err)
		return err
	}

	//插入任务表...
	fmt.Printf("用户记录总数为【%d】,插入类型为[%d] \n ", len(u), t)
	tablename := "wanbu_health_task_rel_user"
	stepth := len(u) / def
	fmt.Printf("分【%d】次插入%s表，每次%d条\n", stepth, tablename, def)

	for i := 0; i < stepth; i++ {

		sqlStr := "insert into " + tablename + " (taskid,userid) values "

		vals := []interface{}{}

		for j := i * def; j < (i+1)*def; j++ {
			sqlStr += "(?,?),"
			vals = append(vals, taskid, u[j].Userid)
		}
		//trim the last ,
		sqlStr = sqlStr[0 : len(sqlStr)-1]
		//format all vals at once
		_, err = c.Db1.Exec(sqlStr, vals...)

		if err != nil {
			Logger.Critical(err)
			return err
		}
		fmt.Printf("总[%d]条数据,总[%d]批,第[%d]批处理完毕,此批[%d]记录\n", len(u), stepth, i, def)
	}

	yu := len(u) % def

	//模除部分处理
	if yu != 0 {

		sqlStr := "insert into " + tablename + " (taskid,userid) values "

		vals := []interface{}{}
		for j := stepth * def; j < len(u); j++ {
			sqlStr += "(?,?),"
			vals = append(vals, taskid, u[j].Userid)
		}

		//trim the last ,
		sqlStr = sqlStr[0 : len(sqlStr)-1]
		//format all vals at once
		_, err := c.Db1.Exec(sqlStr, vals...)

		if err != nil {
			return fmt.Errorf("insert sql err:%s", err.Error())
		}

		fmt.Printf("总[%d]条数据,总[%d]批,第[%d]批处理完毕,此批[%d]记录\n", len(u), stepth, stepth,
			len(u[stepth*def:]))
	}

	return nil
}

func PushChufang(u []*Userinfo, c *Config) error {

	//插入推送表...
	fmt.Printf("用户记录总数为【%d】,插入处方推送消息\n ", len(u))
	tablename := "wanbu_pm_pmlist"
	stepth := len(u) / def
	fmt.Printf("分【%d】次插入%s表，每次%d条\n", stepth, tablename, def)

	for i := 0; i < stepth; i++ {

		sqlStr := "insert into " + tablename + " (fromuserid,touserid,pmtype,new,subject,message,oper2) values "

		vals := []interface{}{}

		for j := i * def; j < (i+1)*def; j++ {
			sqlStr += "(1,?,?,1,?,?,3),"
			vals = append(vals, u[j].Userid, c.Message[0], c.Message[1], c.Message[2])
		}
		//trim the last ,
		sqlStr = sqlStr[0 : len(sqlStr)-1]
		//format all vals at once
		_, err = c.Db1.Exec(sqlStr, vals...)

		if err != nil {
			return err
		}
		fmt.Printf("总[%d]条数据,总[%d]批,第[%d]批处理完毕,此批[%d]记录\n", len(u), stepth, i, def)
	}

	yu := len(u) % def

	//模除部分处理
	if yu != 0 {

		sqlStr := "insert into " + tablename + " (fromuserid,touserid,pmtype,new,subject,message,oper2) values "

		vals := []interface{}{}

		for j := stepth * def; j < len(u); j++ {

			//sqlStr += "(1,?,'%s',1,'%s','%s',3),"
			sqlStr += "(1,?,?,1,?,?,3),"
			vals = append(vals, u[j].Userid, c.Message[0], c.Message[1], c.Message[2])
		}

		//trim the last ,
		sqlStr = sqlStr[0 : len(sqlStr)-1]
		//format all vals at once
		_, err = c.Db1.Exec(sqlStr, vals...)

		if err != nil {
			return fmt.Errorf("insert sql err:%s", err.Error())
		}

		fmt.Printf("总[%d]条数据,总[%d]批,第[%d]批处理完毕,此批[%d]记录\n", len(u), stepth, stepth,
			len(u[stepth*def:]))
	}

	return nil
}

func PushGuding(u []*Userinfo, c *Config) error {

	//插入推送表...
	fmt.Printf("用户记录总数为【%d】,插入固定推送消息\n ", len(u))
	tablename := "wanbu_pm_pmlist"
	stepth := len(u) / def
	fmt.Printf("分【%d】次插入%s表，每次%d条\n", stepth, tablename, def)

	for i := 0; i < stepth; i++ {

		sqlStr := "insert into " + tablename + " (fromuserid,touserid,pmtype,new,subject,message,oper2) values "

		vals := []interface{}{}

		for j := i * def; j < (i+1)*def; j++ {

			sqlStr += "(1,?,?,1,?,?,4),"
			vals = append(vals, u[j].Userid, c.Message[0], c.Message[1], c.Message[2])
		}
		//trim the last ,
		sqlStr = sqlStr[0 : len(sqlStr)-1]
		//format all vals at once
		_, err = c.Db1.Exec(sqlStr, vals...)

		if err != nil {
			return err
		}
		fmt.Printf("总[%d]条数据,总[%d]批,第[%d]批处理完毕,此批[%d]记录\n", len(u), stepth, i, def)
	}

	yu := len(u) % def

	//模除部分处理
	if yu != 0 {

		sqlStr := "insert into " + tablename + " (fromuserid,touserid,pmtype,new,subject,message,oper2) values "

		vals := []interface{}{}

		for j := stepth * def; j < len(u); j++ {

			sqlStr += "(1,?,?,1,?,?,4),"
			vals = append(vals, u[j].Userid, c.Message[0], c.Message[1], c.Message[2])
		}

		//trim the last ,
		sqlStr = sqlStr[0 : len(sqlStr)-1]
		//format all vals at once
		_, err = c.Db1.Exec(sqlStr, vals...)

		if err != nil {
			return fmt.Errorf("insert sql err:%s", err.Error())
		}

		fmt.Printf("总[%d]条数据,总[%d]批,第[%d]批处理完毕,此批[%d]记录\n", len(u), stepth, stepth,
			len(u[stepth*def:]))
	}

	return nil
}
