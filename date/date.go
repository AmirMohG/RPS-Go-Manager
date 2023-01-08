package date
import (
	"fmt"
	"time"
	"strconv"
	"strings"

)
type Date struct {
    Year int
    Month int
    Day int
}
func Difference(date Date, DATE Date) Date{
    result := Date{Year: date.Year-DATE.Year, Month: date.Month-DATE.Month, Day: date.Day-DATE.Day}
    return result
}
func ToString(date Date) string{
    return fmt.Sprintf("%v", date.Year) + " " + fmt.Sprintf("%v", date.Month) + " " + fmt.Sprintf("%v", date.Day)
}
func ToDate(str string) Date {
    split := strings.Split(str, " ")
    var date Date
    Year, _ := strconv.Atoi(split[0])
    Month, _ := strconv.Atoi(split[1])
    Day, _ := strconv.Atoi(split[2])
    date.Year = Year
    date.Month = Month
    date.Day = Day

    return date
}
func Now(date *Date) {
    year, m, day := time.Now().Date()
    var month int = int(m)
    date.Year = year
    date.Month = month
    date.Day = day
}

func IncYear(date *Date, inc int) {
    if inc < 0 {
        
    } else {
        for inc > 0 {
            date.Year++
            inc--
        }
        
    }
    
}
func IncMonth(date *Date, inc int) {
    if inc < 0 {
   
    } else {
        for inc > 0 {
            date.Month++
            inc--
        }
        
    }
    
}
func IncDay(date *Date, inc int) {
    if inc < 0 {
  
    } else {
        for inc > 0 {
            date.Day++
            inc--
        }
        
    }
    
}