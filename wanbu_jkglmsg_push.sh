###########################################################
#                                                         #
# Name   : jkglmsgpush                                    #
#                                                         #
# Usage  :                                                #
#      This script is used to start the WalkDateT1        #
#      application.                                       #
#                                                         #
# Author : Xiejinke                                       #
#                                                         #
# History: 2016-12-05 created                             #
#                                                         #
###########################################################

PROGRAM_HOME="/home/yanglx/WANBU_JKGLMSGPUSH"
MONITOR="$PROGRAM_HOME/bin/wanbu_jkglmsg_push.sh monitor"

cd $PROGRAM_HOME/bin
APP="jkglmsgpush"

start()
{
    list > /dev/null 2>&1
    if [ $? -ne 0 ]
    then
        echo "Start failed. The program had been run !"
    else
        echo ""
        echo "Begin to run $APP ....."
        ./$APP > /dev/null 2>&1 &
        sleep 1
        ps -ef| grep $APP | grep -v grep >/dev/null 2>&1
        if test $? -ne 0
        then
            echo Start $APP unsuccessfully!
        else
            crontab -l>crontab.tmp
            grep "wanbu_jkglmsg_push.sh monitor" crontab.tmp>/dev/null 2>&1
            if test $? -eq 1
            then
                echo "* * * * * . $HOME/.bash_profile; $MONITOR >/dev/null 2>&1">>crontab.tmp
                crontab crontab.tmp
            fi
            rm crontab.tmp
        fi
    fi
}

stop1()
{
    ps -ef| grep $1 | grep -v grep >/dev/null 2>&1
    if test $? -eq 1
    then
        echo Process $1 is not alive !
        else
        proID=`ps -ef| grep $1 |grep -v grep| awk '{ print $2 }'`
        kill -9 $proID 2>/dev/null
        ps -ef| grep $1 | grep -v grep >/dev/null 2>&1
        if test $? -eq 0
        then
            kill -9 $proID 2>/dev/null
        fi
        echo Stop $1 successfully!
       fi
}

stop()
{
    crontab -l | grep -v "wanbu_jkglmsg_push.sh monitor" > crontab.tmp
    crontab crontab.tmp
    rm crontab.tmp

    stop1 $APP

    echo ""
}

about()
{
    if [ $1 ]
    then
        if [ -x $1 ]
        then
                ./$1  -v
        else
            echo "There are no file [$1] here"
        fi
    else
           ./$APP  -v
    fi
    echo ""
}

usage()
{
    echo "Usage:" 
    echo " 查看版本号 $0 about"
    echo " 查看启动情况 $0 list"
    echo " 启动jkglmsgpush $0 start"
    echo " 停止jkglmsgpush $0 stop"
    echo ""
}

list1()
{
    echo "  @$1[$2]:"
    ps -ef| grep $2 | grep -v grep >/dev/null 2>&1
    if test $? -eq 1
    then
        echo "  The process is not alive"
        return 0
    else
        ps -ef | grep $2 |grep -v grep| awk '{print "\t"$1" "$(NF-1)" "$NF}' | sort
        return 1
    fi
}

list()
{
    echo "  >>> The active process(es) as the following:"

    list1 健康管理消息后台推送模块 $APP
}

monitor()
{

    if [ -x $1 ]
    then
        ps -ef| grep $1 | grep -v grep >/dev/null 2>&1
        if test $? -ne 0
        then
            ./$1  2>&1 &
        fi
    fi
}

# See how we were called.
case "$1" in
    start)      start   2>/dev/null;;
    stop)       stop                2>/dev/null;;
    about)      about $APP      2>/dev/null;;
    list)       list                2>/dev/null;;
    monitor)    monitor $APP 2>/dev/null;;
    *)          usage           2>/dev/null;;
esac
exit 0