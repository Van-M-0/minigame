package main

import (
	"fmt"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
)

var db *dbClient

type request struct {
	req 		func()
}

type databseOption struct {
	host 		string
	user 		string
	pwd 		string
	name 		string
	detailLog 	bool
	singular 	bool
}

type dbClient struct {
	opt 		*databseOption
	db 			*gorm.DB
	uri 		string
	reqChan 	chan *request
}
//CREATE DATABASE IF NOT EXISTS mygame default charset utf8 COLLATE utf8_general_ci;
func newDbClient() *dbClient {
	dc := &dbClient{}
	dc.reqChan = make(chan *request, 1024)
	opt := &databseOption{
		host: "127.0.0.1:3306",
		user: "root",
		pwd: "1",
		name: "ergui",
		detailLog: false,
		singular: true,
	}

	uri := fmt.Sprintf("%v:%v@tcp(%v)/%v?charset=utf8&parseTime=True",
		opt.user,
		opt.pwd,
		opt.host,
		opt.name,
	)

	//fmt.Println("db proxy connection info ", uri)
	db, err := gorm.Open("mysql", uri)
	if err != nil {
		fmt.Println("create db proxy err ", err)
		return nil
	}

	if opt.detailLog {
		db.LogMode(true)
	}

	if opt.singular {
		db.SingularTable(true)
	}

	dc.opt = opt
	dc.db = db
	dc.uri = uri
	dc.InitTable()
	dc.handleRequest()
	return dc
}

func (dc *dbClient) handleRequest() {
	go func() {
		for {
			select {
			case r := <- dc.reqChan:
				r.req()
			}
		}
	}()
}

func (dc *dbClient) InitTable() {
	fmt.Println("init tables")
	/*
		dc.DropTable(&table.T_Accounts{})
		dc.DropTable(&table.T_Games{})
		dc.DropTable(&table.T_GamesArchive{})
		dc.DropTable(&table.T_Guests{})
		dc.DropTable(&table.T_Message{})
		dc.DropTable(&table.T_Rooms{})
		dc.DropTable(&table.T_RoomUser{})
		dc.DropTable(&table.T_Users{})
		dc.DropTable(&table.T_MyTest{})
	*/
	dc.CreateTableIfNot(&T_Accounts{})
	dc.CreateTableIfNot(&T_Games{})
	dc.CreateTableIfNot(&T_GamesArchive{})
	dc.CreateTableIfNot(&T_Guests{})
	dc.CreateTableIfNot(&T_Message{})
	dc.CreateTableIfNot(&T_Rooms{})
	dc.CreateTableIfNot(&T_RoomUser{})
	dc.CreateTableIfNot(&T_Users{})
	dc.CreateTableIfNot(&T_MyTest{})
}


func (dc *dbClient) CreateTable(v ...interface{}) {
	dc.db.CreateTable(v...)
}

func (dc *dbClient) CreateTableIfNot(v ...interface{}) {
	for _, m := range v {
		if dc.db.HasTable(m) == false {
			dc.db.CreateTable(m).Set("gorm:table_options", "ENGINE=InnoDB CHARSET=utf8")
		}
	}
}

func (dc *dbClient) CreateTableForce(v...interface{}) {
	dc.db.DropTableIfExists(v...)
	dc.db.CreateTable(v...)
}

func (dc *dbClient) DropTable(v ...interface{}) {
	dc.db.DropTableIfExists(v...)
}

// t_accounts : account info
func (dc *dbClient) GetAccountInfo(account string, accInfo *T_Accounts) bool {
	return dc.db.Where(&T_Accounts{Account: account}).First(accInfo).RowsAffected != 0
}

func (dc *dbClient) AddAccountInfo(accInfo *T_Accounts) bool {
	return dc.db.Create(accInfo).RowsAffected != 0
}

// t_users : user info
func (dc *dbClient) AddUserInfo(userInfo *T_Users) bool {
	fmt.Println("add user info : ", userInfo)
	return dc.db.Create(userInfo).RowsAffected != 0
}

func (dc *dbClient) GetUserInfo(account string, userInfo *T_Users) bool {
	return dc.db.Where("account = ? ", account).
		Select("userid, account, name, sex, headimg, level, exp, coins, gems, roomid").
		Find(&userInfo).
		RowsAffected != 0
}

func (dc *dbClient) GetUserInfoByName(name string, users *T_Users) bool {
	return dc.db.Where("name = ?", name).
		Select("userid, account, name, sex, headimg, level, exp, coins, gems, roomid").
		Find(&users).
		RowsAffected != 0
}

func (dc *dbClient) GetUserInfoByUserid(userid uint32, userInfo *T_Users) bool {
	return dc.db.Where("userid = ? ", userid).
		Select("userid, account, name, sex, headimg, level, exp, coins, gems, roomid").
		Find(&userInfo).
		RowsAffected != 0
}

func (dc *dbClient) ModifyUserInfo(userid uint32, userInfo *T_Users) bool {
	return dc.db.Model(&T_Users{}).
		Where("userid = ?", userid).
		Update(userInfo).
		RowsAffected != 0
}

func (dc *dbClient) GetUserHistoryByUserid(userid uint32, userInfo *T_Users) bool {
	return dc.db.Where("userid = ? ", userid).
		Select("history").
		Find(&userInfo).
		RowsAffected != 0
}

func (dc *dbClient) GetUserGemsByUserid(userid uint32, userInfo *T_Users) bool {
	return dc.db.Where("userid = ? ", userid).
		Select("gems").
		Find(&userInfo).
		RowsAffected != 0
}

func (dc *dbClient) GetUserBaseInfo(userid uint32, userInfo *T_Users) bool {
	return dc.db.Where("userid = ? ", userid).
		Select("name, sex, headimg").
		Find(&userInfo).
		RowsAffected != 0
}

// t_rooms : room info
func (dc *dbClient) GetRoomInfo(roomid string, roomInfo *T_Rooms) bool {
	return dc.db.Where(&T_Rooms{Id: roomid}).First(roomInfo).RowsAffected != 0
}

func dbRequest(fn func()) {
	if db == nil {
		db = newDbClient()
	}
	db.reqChan <- &request{
		req: fn,
	}
}

func dbLobbyUserLogin(account, name, headimg string, sex uint8, cb func(accounts *T_Accounts, users *T_Users, err int)) {
	dbRequest(func() {
		var acc T_Accounts
		var user T_Users
		ok := db.GetAccountInfo(account, &acc)
		db.GetUserInfo(account, &user)
		if ok {
			cb(&acc, &user, 0)
		} else {
			db.AddAccountInfo(&T_Accounts{
				Account: account,
				Password: "123456",
			})
			db.AddUserInfo(&T_Users{
				Account: account,
				Name: name,
				Sex: uint8(sex),
				Headimg: headimg,
			})
			ok = db.GetAccountInfo(account, &acc)
			db.GetUserInfo(account, &user)
			if ok {
				cb(&acc, &user, 0)
			} else {
				cb(nil, nil, 1)
			}
		}
	})
}

func dbCheckConnection() {
	dbRequest(func() {
		fmt.Println("db checking ...")
	})
}

