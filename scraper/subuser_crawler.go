package scraper

import (
	"fmt"
	"math/rand"
	"strconv"
	"strings"
	"time"
	"unicode"

	scraperUtil "github.com/alceccentric/beck-crawler/scraper/util"
	"github.com/gocolly/colly"
)

const (
	subjectCollectionUrlFormat  = "https://bangumi.tv/subject/%s/collections?page=%d"
	collectionTimeFormat        = "2006-1-2 15:04"
	collectionTimeHorizonInDays = 90
)

type SubjectUserScraper struct {
	collyCollector     *colly.Collector
	oldestAccpetedTime time.Time
}

func NewSubjectUserScraper() *SubjectUserScraper {
	return &SubjectUserScraper{
		collyCollector:     initCollector(),
		oldestAccpetedTime: time.Now().AddDate(0, 0, -collectionTimeHorizonInDays),
	}
}

func (c *SubjectUserScraper) Crawl(sid string) {
	fmt.Printf("Crawling subject id: %s\n", sid)
	uids := make([]string, 0)

	c.collyCollector.OnHTML("div.mainWrapper", func(page *colly.HTMLElement) {
		beyondTimeHorizon := false
		page.ForEachWithBreak("li.user", func(_ int, col *colly.HTMLElement) bool {
			uid := col.Attr("data-item-user")
			collectionTime := getCollectionTime(col)

			if c.isBeyondTimeHorizon(collectionTime) {
				beyondTimeHorizon = true
				return false
			} else {
				fmt.Print("Put uid:", uid, "\n")
				uids = append(uids, uid)
				return true
			}
		})

		curIndex := getCurIndex(page)
		maxIndex := getMaxIndex(page)

		if curIndex < maxIndex && !beyondTimeHorizon {
			page.Request.Ctx.Put("maxIndex", strconv.Itoa(maxIndex))
			page.Request.Visit(fmt.Sprintf(subjectCollectionUrlFormat, sid, curIndex+1))
		} else {
			fmt.Printf("Finished crawling subject id: %s\n", sid)
			fmt.Print("curIndex: ", curIndex, " maxIndex: ", maxIndex, " beyondTimeHorizon: ", beyondTimeHorizon, "\n")
		}
	})

	c.collyCollector.OnRequest(func(r *colly.Request) {
		fmt.Println("Visiting", r.URL)
	})

	c.collyCollector.Visit(fmt.Sprintf(subjectCollectionUrlFormat, sid, 1))
	c.collyCollector.Wait()
}

func (c *SubjectUserScraper) isBeyondTimeHorizon(inTime time.Time) bool {
	return inTime.Before(c.oldestAccpetedTime)
}

func initCollector() *colly.Collector {
	agentGen := scraperUtil.NewUserAgentGenerator()
	collector := colly.NewCollector(
		colly.UserAgent(agentGen.RandomUserAgent()),
		colly.Async(true),
	)

	rand.New(rand.NewSource(time.Now().UnixNano()))
	collector.Limit(&colly.LimitRule{
		DomainGlob:  "*",
		Parallelism: 1,
		RandomDelay: time.Duration((rand.Intn(4) + 1)) * time.Second,
	})
	return collector
}

func getCollectionTime(collection *colly.HTMLElement) time.Time {
	pInfoContent := collection.ChildText("p.info")
	if parsedTime, err := time.Parse(collectionTimeFormat, replaceNonASCIIWithSpaces(pInfoContent)); err != nil {
		panic(err)
	} else {
		return parsedTime
	}
}

func getCurIndex(page *colly.HTMLElement) int {
	if curIndex, err := strconv.Atoi(page.ChildText("strong.p_cur")); err != nil {
		panic(err)
	} else {
		return curIndex
	}
}

func getMaxIndex(page *colly.HTMLElement) int {
	if maxIndex := page.Request.Ctx.Get("maxIndex"); maxIndex != "" {
		// has max index in context
		fmt.Printf("Get max index from context: %s\n", maxIndex)
		if mi, err := strconv.Atoi(maxIndex); err != nil {
			panic(err)
		} else {
			return mi
		}
	}

	pEdgeContent := replaceNonASCIIWithSpaces(page.ChildText("span.p_edge"))

	// When p_edge is empty, it means the # of pages is limited
	// Max index can be obtained by iterating through the page anchors
	if pEdgeContent == "" {
		fmt.Println("p_edge is empty, find max index by iterating through the page anchors")
		maxIndex := 1
		page.ForEach("div.page_inner a.p", func(_ int, pageAnchor *colly.HTMLElement) {
			if !unicode.IsDigit([]rune(pageAnchor.Text)[0]) {
				fmt.Printf("Skipping anchor: %s\n", pageAnchor.Text)
				return
			}

			if anchorIndex, err := strconv.Atoi(pageAnchor.Text); err != nil {
				fmt.Printf("Failed to parse index from %s\n", pageAnchor.Text)
				panic(err)
			} else {
				if anchorIndex > maxIndex {
					maxIndex = anchorIndex
				}
			}
		})
		return maxIndex
	} else {
		fmt.Println("Find max index from p_edge")
		if maxIndex, err := strconv.Atoi(strings.Trim(strings.Split(pEdgeContent, "/")[1], " )")); err != nil {
			fmt.Printf("Failed to parse max index from %s\n", pEdgeContent)
			panic(err)
		} else {
			return maxIndex
		}
	}
}

func replaceNonASCIIWithSpaces(input string) string {
	return strings.Map(func(r rune) rune {
		if r > 127 {
			return ' '
		}
		return r
	}, input)
}
