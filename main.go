package main

import (
	"encoding/csv"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

var url string = "https://www.google.com/search?q="
var url2 string = "&rlz=1C5CHFA_enKR933KR933&ei=cwArYIbGEYeOr7wPrrWz2As&start="
var url3 string = "&sa=N&ved=2ahUKEwjGy_y8gu3uAhUHx4sBHa7aDLsQ8tMDegQIDhA6&biw=1112&bih=697"

// 0 10 20

type finalData struct {
	title string
	link  string
}

var errChecking string = "working"

func main() {
	e := echo.New()

	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	e.GET("/", func(c echo.Context) error {
		return c.File("index.html")
	})

	e.POST("/search", handlePost)

	port := os.Getenv("PORT")
	if port == "" {
		port = "1323"
	}

	e.Logger.Fatal(e.Start(":" + port))
}

func handlePost(c echo.Context) error {
	errChecking = "working"
	ch := make(chan []finalData)
	term := c.FormValue("term")
	term = strings.Trim(term, " ")
	term = strings.ToLower(term)

	pages := c.FormValue("pages")
	intPage, err := strconv.Atoi(pages)
	checkErr(err)

	var results []finalData
	for i := 0; i <= intPage; i += 10 {
		fmt.Println("Handling", i)
		go scrape(i, term, ch)
	}
	for i := 0; i <= intPage; i += 10 {
		result := <-ch
		results = append(results, result...)
	}
	writeCsv(results, term)
	defer os.Remove(term + ".csv")
	if errChecking != "working" {
		return c.File("error.html")
	}

	return c.Attachment("./"+term+".csv", term+".csv")
}

func writeCsv(results []finalData, term string) {
	file, err := os.Create(term + ".csv")
	checkErr(err)

	w := csv.NewWriter(file)

	defer w.Flush()

	headers := []string{"Title", "Link"}

	Err := w.Write(headers)
	checkErr(Err)

	for _, data := range results {
		resultsSlice := []string{data.title, data.link}
		Err2 := w.Write(resultsSlice)
		checkErr(Err2)
	}
}

func scrape(i int, term string, ch chan []finalData) {
	var results []finalData
	req, err := http.NewRequest("GET", url+term+url2+strconv.Itoa(i)+url3, nil)
	checkErr(err)
	req.Header.Add("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 11_1_0) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/87.0.4280.88 Safari/537.36")
	res, err := http.DefaultClient.Do(req)
	checkErr(err)
	checkStatus(res.StatusCode)
	doc, err := goquery.NewDocumentFromReader(res.Body)
	checkErr(err)

	doc.Find(".tF2Cxc").Each(func(i int, s *goquery.Selection) {
		title := s.Find(".LC20lb").Text()
		link, _ := s.Find(".yuRUbf a").Attr("href")
		result := finalData{title: title, link: link}
		results = append(results, result)
	})
	ch <- results
}

func checkErr(err error) {
	if err != nil {
		errChecking = "not working"
	}
}

func checkStatus(code int) {
	if code != 200 {
		fmt.Println("Status is not 200", code)
		errChecking = "not working"
	}
}
