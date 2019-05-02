package mariadb

import (
	"fmt"
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
	// "log"
)

func Conn2db() *sql.DB {
	stmt := fmt.Sprintf("%s:%s@/%s?parseTime=true&charset=utf8",DBUser,DBPass,DBName)
	db, _ := sql.Open("mysql",stmt)
	err := db.Ping()
	if err != nil {
		fmt.Println(err)
	}
	var version string
	db.QueryRow("SELECT VERSION()").Scan(&version)
	fmt.Println("Connected to:", version)
	return db
}

var db = Conn2db()

func InsertPair(user string,userMatchTarget string) {
	stmt := fmt.Sprintf("INSERT INTO %s%s VALUES(?,?)",PairTable,"(user,userMatchTarget)")
	println(stmt)
	stmtInsPair,err := db.Prepare(stmt)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer stmtInsPair.Close()
	_,err = stmtInsPair.Exec(user,userMatchTarget)
	if err != nil {
		println(err)
		return
	}
	println("Insert into pair done!")
}

func UpdatePair(user string,userMatchTarget string) {
	stmt := fmt.Sprintf("UPDATE %s SET %s=? WHERE %s=?",PairTable,userMatchTarget,user)
	println(stmt)
	stmtInsPair,err := db.Prepare(stmt)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer stmtInsPair.Close()
	_,err = stmtInsPair.Exec(user,userMatchTarget)
	if err != nil {
		println(err)
		return
	}
	println("Update pair done!")
}

func InsertMsg(openId string,answer string) {
	stmt := fmt.Sprintf("INSERT INTO %s VALUES(?,?)",MsgTable+"(_from,_to,msgType,msgContent)")
	println(stmt)
	stmtInsMsg,err := db.Prepare(stmt)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	defer stmtInsMsg.Close()
	_,err = stmtInsMsg.Exec(openId,answer)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	println("Insert into msg done!")
}

func InsertUser(openId string,answer int64) {
	stmt := fmt.Sprintf("INSERT INTO %s VALUES(?,?)",UserTable)
	// println(stmt)
	stmtInsUser,err := db.Prepare(stmt)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer stmtInsUser.Close()
	_,err = stmtInsUser.Exec(openId,answer)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	println("Insert into user done!")
}

func UpdateUser(openId string,answer int64) {
	stmt := fmt.Sprintf("UPDATE %s SET %s=? WHERE %s=?",UserTable,answer,openId)
	// println(stmt)
	stmtInsUser,err := db.Prepare(stmt)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer stmtInsUser.Close()
	_,err = stmtInsUser.Exec(answer,openId)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	println("Update user done!")
}

func QueryIfUserExist(openId string) bool{
	isExist := false
	stmt := fmt.Sprintf("select openId from %s",UserTable)
	rows,err := db.Query(stmt)
	if err != nil {
		fmt.Println(err.Error())
		panic("Query user error in `QueryUser`")
	}
	defer rows.Close()
	var fetchUser string
	for rows.Next() {
		err := rows.Scan(&fetchUser)
		if err != nil {
			fmt.Println(err.Error())
			panic("Fetch user error in `QueryUser`")
		}
		if openId == fetchUser {
			isExist = true
			break
		}
	}
	return isExist
}

func QueryPairGetTarget(openId string) string{
	var queryResult string
	var (
		candidateUser string
		candidateUserMatchTarget string
	)
	stmt := fmt.Sprintf("select user,userMatchTarget from %s",PairTable)
	rows,err := db.Query(stmt)
	if err != nil {
		fmt.Println(err.Error())
		panic("Query Pair error in `QueryPair`")
	}
	defer rows.Close()
	for rows.Next() {
		err := rows.Scan(&candidateUser,&candidateUserMatchTarget)
		if err != nil {
			fmt.Println(err.Error())
			return ""
		}
		if openId == candidateUser {
			queryResult = candidateUserMatchTarget
			break
		}
	}
	return queryResult
}

func SimilarWeight(ans1 int64,ans2 int64,quizNumber int) int{
	var weight int
	var mask int
	mask = 1
	for mask<quizNumber {
		if ((1<<uint(mask)) & ans1) == ((1<<uint(mask))&ans2) {
			weight += 1
		}
		mask++
	}
	return weight
}

func MakePair(openId string,answer int64) string{
	var curWeight int
	var curTarget string
	var(
		id string
		ans int64
	)
	stmt := fmt.Sprintf("select openId,answer from %s",UserTable)
	rows,err := db.Query(stmt)
	if err != nil {
		println("Query user fail in MakePair")
		fmt.Println(err.Error())
		return "unknow"
	}
	for rows.Next() {
		err := rows.Scan(&id,&ans)
		if err != nil {
			println("Fetch user fail in MakePair")
			fmt.Println(err.Error())
			return "unknow"
		}
		if openId != id {
			tempWeight := SimilarWeight(answer,ans,QuizNumber)
			if tempWeight > curWeight {
				curTarget = id
			}
		}
	}
	if curTarget == "" {
		println("only 1 user in database")
		return "unique"
	} else {
		InsertPair(openId,curTarget)
		return "registerOk"
	}
}

func MakeNewPair(openId string,answer int64) string{
	UpdateUser(openId,answer)
	var curWeight int
	var curTarget string
	var(
		id string
		ans int64
	)
	stmt := fmt.Sprintf("select openId,answer from %s",UserTable)
	rows,err := db.Query(stmt)
	if err != nil {
		println("Query user fail in MakeNewPair")
		fmt.Println(err.Error())
		return "unknow"
	}
	for rows.Next() {
		err := rows.Scan(&id,&ans)
		if err != nil {
			println("Fetch user fail in MakeNewPair")
			fmt.Println(err.Error())
			return "unknow"
		}
		if openId != id {
			tempWeight := SimilarWeight(answer,ans,QuizNumber)
			if tempWeight > curWeight {
				curTarget = id
			}
		}
	}
	if curTarget == "" {
		// println("only 1 user in database")
		return "unknow"
	} else {
		UpdatePair(openId,curTarget)
		return "registerOk"
	}
}