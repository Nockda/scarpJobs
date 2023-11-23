package main

import (
	"encoding/csv"
	"fmt"
	"goquery"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
)

type extractedJob struct{
	id 		string
	title 	string
	location 	string
	types string
}

var baseURL string = "https://www.saramin.co.kr/zf_user/search/recruit?&searchword=spark"

func main() {
	var jobs[]extractedJob
	c := make(chan []extractedJob)
	totalPages := getPages()

	for i := 0; i<totalPages; i++{
		go getPage(i, c)

	}
	for i := 0; i<totalPages; i++{
		extractJobs := <-c
		jobs = append(jobs, extractJobs...)
	}
	writeJobs(jobs)
	fmt.Println("It's done. Length was ", len(jobs))
}


func getPage(page int, mainC chan <- []extractedJob){
	var jobs [] extractedJob
	c := make(chan extractedJob)
	pageURL := baseURL+"&recruitPage="+ strconv.Itoa(page)
	fmt.Println("Requesting", pageURL)
	res, err :=http.Get(pageURL)
	checkErr(err)
	checkCode(res)
	defer res.Body.Close()

	doc, err := goquery.NewDocumentFromReader(res.Body)

	checkErr(err)
	searchCards := doc.Find(".item_recruit")
	searchCards.Each(func(i int, card *goquery.Selection){
		go extractJobs(card, c)
	})
	for i :=0; i<searchCards.Length(); i++{
		job := <-c
		jobs = append(jobs, job)
	}
	mainC <- jobs
}

func extractJobs(card *goquery.Selection, c chan<- extractedJob) {
	id, _ := card.Attr("value")
	title := cleanString(card.Find(".corp_name>a").Text())
	locationSelection := card.Find(".job_condition span a")
	var locations []string
	locationSelection.Each(func(i int, location *goquery.Selection) {
		locations = append(locations, location.Text())
	})
	location := strings.Join(locations, " ")
	types := card.Find(".job_condition span:nth-child(4)").Text()
	c <- extractedJob {id: id, 
						title: title, 
						location: location,
		 				types: types}
}

func getPages() int{
	pages := 0
	res, err := http.Get(baseURL)
	checkErr(err)
	checkCode(res)
	

	defer res.Body.Close()

	doc, err := goquery.NewDocumentFromReader(res.Body)

	checkErr(err)
	doc.Find(".pagination").Each(func(i int, s *goquery.Selection){
		pages = s.Find("a").Length()
	})
	return pages
}

func checkErr (err error){
	if err!= nil{
		log.Fatalln(err)
	}
}

func writeJobs(jobs[]extractedJob){
	file, err := os.Create("jobs.csv")
	checkErr(err)

	w := csv.NewWriter(file)
	defer w.Flush()

	headers := []string{"ID","Title", "Location", "Types"}
	wErr := w.Write(headers)
	checkErr(wErr)

	for _, job := range jobs {
		jobSlice := []string{job.id, job.title, job.location, job.types}
		jwErr := w.Write(jobSlice)
		checkErr(jwErr)
	}
}

func checkCode(res * http.Response){
	if res.StatusCode!=200{
		log.Fatalln("Request failed with Status: " , res.StatusCode)
	}
}


func cleanString(str string) string{
	return strings.Join(strings.Fields(strings.TrimSpace(str)), " ")
}