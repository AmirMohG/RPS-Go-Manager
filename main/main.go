package main

import (
    "fmt"
    "golang.org/x/crypto/bcrypt"
    "database/sql"
    "github.com/joho/godotenv"
    "os"
    "os/exec"
    _ "time"
    _ "github.com/go-sql-driver/mysql"
    "github.com/gin-gonic/gin"
    _ "net/http"
    _ "io"
    _ "strings"
    _ "strconv"
    "rps.local/date"
)

type reqBody struct {
    Username string `json:"username"`
    Password string `json:"password"`
    Vmid string `json:"vmid"`
    Expiry date.Date `json:"expiry"`
    Ram float32 `json:"ram"`
    Cpu float32 `json:"cpu"`
    Price int `json:"price"`
}
type User struct {
    Username string `json:"username"`
    Password string `json:"password"`
    Date_created date.Date `json:"date_created"`
}
 
type VM struct {
    Vmid string `json:"vmid"`
    Owner string `json:"owner"`
    Expiry date.Date `json:"expiry"`
    Ram float32 `json:"ram"`
    Cpu float32 `json:"cpu"`
    Price int `json:"price"`
}
var db *sql.DB
func main() {
    godotenv.Load(".env")
    APP_HOST := os.Getenv("APP_HOST")
    APP_PORT := os.Getenv("APP_PORT")
    ADDR := fmt.Sprintf(APP_HOST + ":" + APP_PORT)
    fmt.Println(ADDR)
    MYSQL_HOST := os.Getenv("MYSQL_HOST")
    MYSQL_DB := os.Getenv("MYSQL_DB")
    MYSQL_USER := os.Getenv("MYSQL_USER")
    MYSQL_PASS := os.Getenv("MYSQL_PASS")
    MYSQL_PORT := os.Getenv("MYSQL_PORT")
    db = databaseCon(fmt.Sprintf(MYSQL_USER + ":" + MYSQL_PASS +"@tcp(" + MYSQL_HOST + ":" + MYSQL_PORT + ")/" + MYSQL_DB))
    // vm := VM{vmid: "203", owner: "amg", expiry: 25, ram: 2, cpu: 2}
    var d date.Date
    date.Now(&d)
    fmt.Println(date.ToString(d))
    fmt.Println(date.ToString(date.ToDate("2023 1 9")))
    
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
            c.JSON(200, gin.H{"StatusCode": register(user, )})

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
        
            c.JSON(200, gin.H{"StatusCode": buyVM(user, vm)})

    })
	router.Run(ADDR)
    //fmt.Println(listVMs("amg"))
    //fmt.Println(authorized(user, vm))
    //buyVM("amg", "205", 1, 1, 55)
    //restartVM("amg", "205")
    //fmt.Println(userExists(user))
    //register("user4", "w20805338")


}
func databaseCon(URI string) *sql.DB{
    db , err := sql.Open("mysql", URI)
    if err != nil {
        panic(err)
    }
    return db;
}
func register(user User) string {
    if !userExists(user) && len(user.Username) < 32 && len(user.Username) > 3 && len(user.Password) > 8 {
        hashedPass := hash(user)
        query := fmt.Sprintf("INSERT INTO Users (username, password, date_created) VALUES (\"" + user.Username + "\", \"" + hashedPass + "\", " + "NOW()" + ")")
        _, err := db.Exec(query)
        if err != nil {
            panic(err)
        }
    } else {
        return "Bad credentials"
    }
    return "200"
}


func userExists(user User) bool {
    query := fmt.Sprintf("select username from Users where username = \"" + user.Username + "\"")
    err := db.QueryRow(query).Scan(&user.Username)
    if err != nil {
	    return false
    } else {
        return true
    }
    }
func hash(user User) string {
    password := []byte(user.Password)

    // Hashing the password with the default cost of 10
    hashedPassword, err := bcrypt.GenerateFromPassword(password, bcrypt.DefaultCost)
    if err != nil {
        panic(err)
    }
    return string(hashedPassword[:])

}
func buyVM (user User, vm VM) string {
    //|| vm.Expiry < time.Now() 
    if vm.Vmid == "" || vm.Owner == "" || vm.Ram < 100 || vm.Cpu < 100 || vm.Price < 5 {
        return "Bad credentials"
    } else if paid() {
        query := fmt.Sprintf("INSERT INTO VMs (vmid, owner, expiry, ram, cpu, price) VALUES (\"" + vm.Vmid + "\", \"" + user.Username + "\", " + "NOW()" + ", " + fmt.Sprintf("%f", vm.Ram) + ", " + fmt.Sprintf("%f", vm.Cpu) + ", " + fmt.Sprintf("%v", vm.Price) + ")")
        
        _, err := db.Exec(query)
        fmt.Println(query)
        if err != nil {
            return "500"
        } else {
            return "200"
        }
    }
    return "401"
}

func listVMs(user User) []string {
    var vm VM
    vms := make([]string, 1)
    query := fmt.Sprintf("select vmid from VMs where owner = \"" + user.Username + "\"")
    rows, err := db.Query(query)
    if err != nil {
        panic(err)
    }
    defer rows.Close()
    for rows.Next() {
        err := rows.Scan(&vm.Vmid)
        vms = append(vms, vm.Vmid)
        if err != nil {
            panic(err)
        }
    }
    err = rows.Err()
    if err != nil {
        panic(err)
    }

    return vms
}
func authorized(user User, vm VM) bool {
    query := fmt.Sprintf("select vmid from VMs where owner = \"" + user.Username + "\"")
    vmid := vm.Vmid
    rows, err := db.Query(query)
    if err != nil {
        return false
    }
    defer rows.Close()
    for rows.Next() {
        err := rows.Scan(&vm.Vmid)
        if vm.Vmid == vmid {
            return true
        }
        if err != nil {
            panic(err)
        }
    }
    err = rows.Err()
    if err != nil {
        panic(err)
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