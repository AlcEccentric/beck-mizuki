package scraper

import (
	"fmt"
	"math/rand"
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/alceccentric/beck-crawler/util"
	"github.com/gocolly/colly"
)

type SubjectUserScraper struct {
	collector          *colly.Collector
	oldestAccpetedTime time.Time
	uidChan            chan string
}

func NewSubjectUserScraper(coldStartIntervalInDays, uidChanSize int) *SubjectUserScraper {
	subjectUserScraper := &SubjectUserScraper{
		collector:          initColly(),
		oldestAccpetedTime: time.Now().AddDate(0, 0, -coldStartIntervalInDays),
		uidChan:            make(chan string, uidChanSize),
	}
	subjectUserScraper.registerHandler()
	return subjectUserScraper
}

func initColly() *colly.Collector {
	agentGen := NewUserAgentGenerator()
	collector := colly.NewCollector(
		colly.UserAgent(agentGen.RandomUserAgent()),
		colly.Async(true),
	)

	randGen := rand.New(rand.NewSource(time.Now().UnixNano()))
	collector.Limit(&colly.LimitRule{
		DomainGlob:  "*",
		Parallelism: 1,
		RandomDelay: time.Duration((randGen.Intn(3) + 2)) * time.Second,
	})
	return collector
}

func (scraper *SubjectUserScraper) Crawl(sid string) {
	ctx := colly.NewContext()
	ctx.Put("subjectId", sid)

	scraper.collector.Request("GET", fmt.Sprintf(util.SubjectCollectionUrlFormat, sid, 1), nil, ctx, nil)
	scraper.collector.Wait()
}

func (scraper *SubjectUserScraper) CloseUidChan() {
	close(scraper.uidChan)
}

func (scraper *SubjectUserScraper) CollectUids() []string {
	uidSet := make(map[string]struct{})
	for uid := range scraper.uidChan {
		uidSet[uid] = struct{}{}
	}

	uidSlice := make([]string, 0, len(uidSet))
	for uid := range uidSet {
		uidSlice = append(uidSlice, uid)
	}
	return uidSlice
}

func (scraper *SubjectUserScraper) registerHandler() {
	scraper.collector.OnHTML("div.mainWrapper", scraper.handleMainWrapper)
}

func (scraper *SubjectUserScraper) handleMainWrapper(page *colly.HTMLElement) {
	maxIndex, err := scraper.getMaxIndex(page)
	fmt.Printf("Processing page %s\n", page.Request.URL.String())
	if err != nil {
		fmt.Println("Failed to get max index:", err)
		return
	}

	sid := page.Request.Ctx.Get("subjectId")
	if sid == "" {
		fmt.Println("No subject id in context")
		return
	}

	beyondTimeHorizon := scraper.processUserCollections(page, sid)
	if !beyondTimeHorizon {
		scraper.checkAndVisitNextPage(page, sid, maxIndex)
	} else {
		// fmt.Printf("Wont check next page and stop at %s\n", page.Request.URL.String())
	}
}

func (scraper *SubjectUserScraper) processUserCollections(page *colly.HTMLElement, sid string) bool {
	beyondTimeHorizon := false
	page.ForEachWithBreak("li.user", func(_ int, col *colly.HTMLElement) bool {
		uid := col.Attr("data-item-user")
		collectionTime, err := scraper.getCollectionTime(col)

		if err != nil {
			fmt.Println("Failed to get collection time for uid:", uid, "for subject id:", sid, "with err:", err)
			return true // skip this and continue
		}

		if scraper.isBeyondTimeHorizon(collectionTime) {
			beyondTimeHorizon = true
			return false // stop processing
		} else {
			scraper.uidChan <- uid
			return true // skip this and continue
		}
	})
	return beyondTimeHorizon
}

func (scraper *SubjectUserScraper) checkAndVisitNextPage(page *colly.HTMLElement, sid string, maxIndex int) {
	curIndex, err := scraper.getCurIndex(page)
	if err != nil {
		fmt.Printf("Error getting current index: %v\n", err)
		return
	}

	if curIndex < maxIndex {
		page.Request.Ctx.Put("maxIndex", strconv.Itoa(maxIndex))
		fmt.Printf("Going to visit subject id: %s, index: %d\n", sid, curIndex+1)
		page.Request.Visit(fmt.Sprintf(util.SubjectCollectionUrlFormat, sid, curIndex+1))
	}
}

func (scraper *SubjectUserScraper) isBeyondTimeHorizon(inTime time.Time) bool {
	return inTime.Before(scraper.oldestAccpetedTime)
}

func (scraper *SubjectUserScraper) getCollectionTime(collection *colly.HTMLElement) (time.Time, error) {
	pInfoContent := collection.ChildText("p.info")
	if parsedTime, err := time.Parse(util.WebsiteCollectionTimeFormat, replaceNonASCIIWithSpaces(pInfoContent)); err != nil {
		return time.Now(), fmt.Errorf("invalid collection time: %s error: %s", pInfoContent, err)
	} else {
		return parsedTime, nil
	}
}

func (scraper *SubjectUserScraper) getCurIndex(page *colly.HTMLElement) (int, error) {
	if curIndex := page.Request.Ctx.Get("curIndex"); curIndex != "" {
		if ci, err := strconv.Atoi(curIndex); err != nil {
			return 0, fmt.Errorf("invalid curIndex: %s  error: %s", curIndex, err)
		} else {
			return ci, nil
		}
	} else {
		return 1, nil
	}
}

func (scraper *SubjectUserScraper) getMaxIndex(page *colly.HTMLElement) (int, error) {
	if maxIndex := page.Request.Ctx.Get("maxIndex"); maxIndex != "" {
		if mi, err := strconv.Atoi(maxIndex); err != nil {
			return 0, fmt.Errorf("invalid maxIndex in context: %s  error: %s", maxIndex, err)
		} else {
			return mi, nil
		}
	}

	pEdgeContent := replaceNonASCIIWithSpaces(page.ChildText("span.p_edge"))

	// When p_edge is empty, it means the # of pages is limited
	// Max index can be obtained by iterating through the page anchors
	if pEdgeContent == "" {
		maxIndex := 1
		page.ForEach("div.page_inner a.p", func(_ int, pageAnchor *colly.HTMLElement) {
			if !unicode.IsDigit([]rune(pageAnchor.Text)[0]) {
				return
			}

			if anchorIndex, err := strconv.Atoi(pageAnchor.Text); err != nil {
				fmt.Printf("failed to parse anchor index: %s\n", pageAnchor.Text)
			} else {
				if anchorIndex > maxIndex {
					maxIndex = anchorIndex
				}
			}
		})
		return maxIndex, nil
	} else {
		if maxIndex, err := strconv.Atoi(strings.Trim(strings.Split(pEdgeContent, "/")[1], " )")); err != nil {
			return 0, fmt.Errorf("invalid p_edge: %s  error: %s", pEdgeContent, err)
		} else {
			return maxIndex, nil
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
