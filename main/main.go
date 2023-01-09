package main

import (
    "fmt"
    "golang.org/x/crypto/bcrypt"
    _ "database/sql"
    "github.com/joho/godotenv"
    "os"
    "os/exec"
    _ "time"
    _ "github.com/go-sql-driver/mysql"
    "github.com/gin-gonic/gin"
    _ "net/http"
    _ "io"
    "strings"
    _ "strconv"
    "rps.local/date"
    "gorm.io/gorm"
    "gorm.io/driver/mysql"
)
type DB struct{
    Username string
    Password string
    Host string
    Port string
    Database string
}
type reqBody struct {
    Username string `json:"username"`
    Password string `json:"password"`
    Vmid string `json:"vmid"`
    Duration int `json:"duration"`
    Ram float32 `json:"ram"`
    Cpu float32 `json:"cpu"`
    Price int `json:"price"`
}
type User struct {
    Username string `json:"username" gorm:"primaryKey" gorm:"column:username;not_null;type:varchar;size:32;unique_index"`
    Password string `json:"password" gorm:"column:password;not null;size:64"`
    Date_created string `json:"date_created" gorm:"column:date_created;not null;size:20"`
}

type VM struct {
    Vmid string `json:"vmid" gorm:"primaryKey" gorm:"column:vmid;not null`
    Owner string `json:"owner" gorm:"column:owner;not null;references:users(username)"`
    Expiry string `json:"expiry" gorm:"column:expiry;not null"`
    Ram float32 `json:"ram" gorm:"column:ram;not null"`
    Cpu float32 `json:"cpu" gorm:"column:cpu;not null"`
    Price int `json:"price" gorm:"column:price;not null"`
    user []User `gorm:"foreignKey:Username;references:Owner"`

}
var db *gorm.DB
func main() {
    godotenv.Load(".env")
    APP_HOST := os.Getenv("APP_HOST")
    APP_PORT := os.Getenv("APP_PORT")
    ADDR := fmt.Sprintf(APP_HOST + ":" + APP_PORT)
    db = InitGormMysql(db)
    
    db.AutoMigrate(&VM{})
    db.AutoMigrate(&User{})
    router := gin.Default()
	router.GET("/register", func(res *gin.Context) {
		res.JSON(200, gin.H{
			"user": "pong",
		})
	})
    router.POST("/api/userExists", func(c *gin.Context) {
        var user User
        err := c.Bind(&user)
        if err != nil {
        panic(err)
    }
        if userExists(user) {
            c.JSON(200, gin.H{"status": "True"})
        } else {
            c.JSON(200, gin.H{"status": "False"})
        }
    })
    router.POST("/api/register", func(c *gin.Context) {
        var user User
        err := c.Bind(&user)
        if err != nil {
            panic(err)
        }
            c.JSON(200, gin.H{"StatusCode": register(user)})

    })
    router.POST("/api/buy", func(c *gin.Context) {
        var user User
        var vm VM
        var req reqBody
        err := c.Bind(&req)
        if err != nil {
            panic(err)
        }
        user.Username = req.Username
        user.Password = req.Password
        vm.Vmid = req.Vmid
        vm.Owner = req.Username
        vm.Ram = req.Ram
        vm.Cpu = req.Cpu
        vm.Price = req.Price
        duration := req.Duration
            c.JSON(200, gin.H{"StatusCode": buyVM(user, vm, duration)})

    })
	router.Run(ADDR)
    

}
func InitGormMysql(conn *gorm.DB) (*gorm.DB) {
	var err error
    dbConf := DB{
        Username: os.Getenv("MYSQL_USER"),
        Password: os.Getenv("MYSQL_PASS"),
        Host: os.Getenv("MYSQL_HOST"),
        Port: os.Getenv("MYSQL_PORT"),
        Database: os.Getenv("MYSQL_DB"),
    }
    fmt.Println("CONFIG", dbConf)
	conn, err = gorm.Open(mysql.New(mysql.Config{
        DSN: fmt.Sprintf("%s:%s@(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
			dbConf.Username,
			dbConf.Password,
			dbConf.Host,
			dbConf.Port,
			dbConf.Database,
		), 
        DefaultStringSize: 256, // default size for string fields
        DisableDatetimePrecision: true, // disable datetime precision, which not supported before MySQL 5.6
        DontSupportRenameIndex: true, // drop & create when rename index, rename index not supported before MySQL 5.7, MariaDB
        DontSupportRenameColumn: true, // `change` when rename column, rename column not supported before MySQL 8, MariaDB
        SkipInitializeWithVersion: false, // auto configure based on currently MySQL version
        }))
	if err != nil {
		panic(err)
	}
    return conn
}
func register(user User) string {
    if !userExists(user) && len(user.Username) < 32 && len(user.Username) > 3 && len(user.Password) > 8 {
        hashedPass := hash(user)
        user.Username = strings.ToLower(user.Username)
        user.Password = hashedPass
        var date_created date.Date
        date.Now(&date_created)
        user.Date_created = date.ToString(date_created)
        res := db.Create(&user)
        if res.Error != nil {
            panic(res.Error)
        }
    } else {
        return "Bad credentials"
    }
    return "200"
    
}

func userExists(user User) bool {
    
    res := db.First(&user, "username = ?", strings.ToLower(user.Username))
    fmt.Println()
    if res.Error != nil {
	    return false
    } else {
        return true
    }
}
func hash(user User) string {
    password := []byte(user.Password)

    hashedPassword, err := bcrypt.GenerateFromPassword(password, bcrypt.DefaultCost)
    if err != nil {
        panic(err)
    }
    return string(hashedPassword[:])
}
func buyVM (user User, vm VM, duration int) string {
    if vm.Vmid == "" || !userExists(user) || vm.Ram < 100 || vm.Cpu < 1 || vm.Price < 5 {
        return "Bad credentials"
    } else if paid() {
        var expiry date.Date
        date.Now(&expiry)
        date.IncMonth(&expiry, duration)
        vm.Expiry = date.ToString(expiry)
        res := db.Create(vm)
        if res.Error != nil {
            panic(res.Error)
        }
        if res.Error != nil {
            return "500"
        } else {
            return "200"
        }
    }
    return "401"
}

func authorized(user User, vm VM) bool {
    if !userExists(user) {
        return false
    }
    res := db.First(&vm, "owner = ?", strings.ToLower(user.Username))
    if res.Error != nil {
	    return false
    } else {
        return true
    }
   
    return false
}
func paid() bool {
    return true
}
func restartVM(user User, vm VM) string{
    if authorized(user, vm) {
        res, err := exec.Command("qm restart "+ vm.Vmid).Output()

        if err != nil {
            panic(err)
            return err.Error()
        }
        return string(res[:])
    } else {
        return "401"
    }
}