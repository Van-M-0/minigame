package main

type T_MyTest struct {
	Account 	string 		`gorm:"size:20;type:char(20)"`
	Name 		string
	Status 		int
}

type T_Accounts struct {
	Account 	string		`gorm:"primary_key"`
	Password	string
}

type T_Games struct {
	RoomUuid	string		`gorm:"size:20;type:char(20);primary_key"`
	GameIndex 	int			`gorm:"type:int unsigned;primary_key"`
	BaseInfo 	string		`gorm:"size:1024;varchar(1024);not null"`
	CreateTime 	int			`gorm:"not null"`
	Snapshots	string 		`gorm:"size:2048;varchar(2048);default:null"`
	ActionRecords string 	`gorm:"varchar(2048);default:null"`
	Result 		string 		`gorm:"default:null"`
}

type T_GamesArchive struct {
	RoomUuid	string		`gorm:"size:20;type:char(20);primary_key"`
	GameIndex 	int			`gorm:"type:int unsigned;primary_key"`
	BaseInfo 	string		`gorm:"size:1024;varchar(1024);not null"`
	CreateTime 	int			`gorm:"not null"`
	Snapshots	string 		`gorm:"size:2048;varchar(2048);default:null"`
	ActionRecords string 	`gorm:"varchar(2048);default:null"`
	Result 		string 		`gorm:"default:null"`
}

type T_Guests struct {
	GuestAccount string 	`gorm:"size:255;type:varchar(255);primary_key"`
}

type T_Message struct {
	Type 		string 		`gorm:"size:32;varchar(32);not null;primary_key"`
	Msg 		string 		`gorm:"size:1024;varchar(1024);not null"`
	Version 	string 		`gorm:"size:32;varchar(32);not null"`
}

type T_Rooms struct {
	Uuid 		string 		`gorm:"size:20;type:char(20);not null;primary_key"`
	Id 			string 		`gorm:"size:8;not null"`
	BaseInfo 	string 		`gorm:"size:256;type:varchar(256);not null;default:0"`
	CreateTime 	int 		`gorm:"not null"`
	NumOfTurns 	int 		`gorm:"not null;default:'0'"`
	NextButton	int 		`gorm:"not null;default:'0'"`
	User1		int			`gorm:"not null"`
	User2		int			`gorm:"not null"`
	User3		int			`gorm:"not null"`
	User4		int			`gorm:"not null"`
}

type T_RoomUser struct {
	UserIndex 	int 		`gorm:"primary_key;AUTO_INCREMENT"`
	UserID 		int 		`gorm:"not null"`
	UserIcon 	string 		`gorm:"size:128;not null;default:''"`
	UserName 	string 		`gorm:"size:32;not null;default:''"`
	UserScore 	int 		`gorm:"not null;default:0"`
}

type T_Users struct {
	Userid 		uint32		`gorm:"primary_key;AUTO_INCREMENT;not null"`
	Account 	string 		`gorm:"size:32;not null;default:'';index:acc_index"`
	Name 		string 		`gorm:"size:32;default:null"`
	Sex 		uint8
	Headimg 	string
	Level 		uint8 		`gorm:"default:1"`
	Exp 		uint32		`gorm:"default:0"`
	Coins 		uint32		`gorm:"default:0"`
	Gems 		uint32 		`gorm:"default:0"`
	Roomid 		string 		`gorm:"size:8"`
	History 	string 		`gorm:"size:4096;not null;default:''"`
}
