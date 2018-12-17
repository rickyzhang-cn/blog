package service

import (
	"html/template"
	"io/ioutil"
	"net/http"
	"strings"

	"encoding/json"

	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/russross/blackfriday"
)

type Index struct {
	Posts []*Post `json:"posts"`
}

type Post struct {
	Title    string   `json:"title"`
	FileName string   `json:"file_name"`
	Date     string   `json:"date"`
	Category string   `json:"category"`
	Tags     []string `json:"tags"`
}

func ParseIndex() gin.HandlerFunc {
	return func(c *gin.Context) {
		content, err := ioutil.ReadFile("./index.json")
		if err != nil {
			c.Abort()
		}
		info := &Index{}
		err = json.Unmarshal(content, info)
		if err != nil {
			c.Abort()
		}
		c.Set("Posts", info.Posts)
		categoryM := make(map[string][]*Post, 16)
		tagM := make(map[string][]*Post, 16)
		for _, post := range info.Posts {
			cate, tags := post.Category, post.Tags
			categoryM[cate] = append(categoryM[cate], post)
			for _, tag := range tags {
				tagM[tag] = append(tagM[tag], post)
			}
		}
		c.Set("CategoryMap", categoryM)
		c.Set("TagMap", tagM)
		categories := make([]string, 0, 16)
		tags := make([]string, 0, 16)
		for cate := range categoryM {
			categories = append(categories, cate)
		}
		for tag := range tagM {
			tags = append(tags, tag)
		}
		c.Set("Categories", categories)
		c.Set("Tags", tags)
		c.Next()
	}
}

func Home(c *gin.Context) {
	info, exist := c.Get("Posts")
	if !exist {
		c.Abort()
	}
	posts := info.([]*Post)
	categories := c.GetStringSlice("Categories")
	tags := c.GetStringSlice("Tags")
	fmt.Println(categories, tags)
	c.HTML(http.StatusOK, "home.html", gin.H{
		"Posts":      posts,
		"Categories": categories,
		"Tags":       tags,
	})
}

func About(c *gin.Context) {

}

func GetCategory(c *gin.Context) {

}

func GetPost(c *gin.Context) {
	path := "./post/" + c.Param("name")
	content, err := ioutil.ReadFile(path)
	if err != nil {
		c.Abort()
	}
	fmt.Println(c.Get("Posts"))
	fmt.Println(c.Get("Tag"))
	fmt.Println(c.Get("Category"))
	title, date, tags := getMetaInfo(string(content))
	html := blackfriday.MarkdownCommon(content)
	c.HTML(http.StatusOK, "post.html", gin.H{
		"Title":   title,
		"Date":    date,
		"Tags":    tags,
		"Content": template.HTML(html),
	})

}

func getMetaInfo(content string) (title string, date string, tags []string) {
	lines := strings.Split(content, "\n")
	if len(lines) <= 4 {
		return
	}
	title = strings.Split(lines[1], ":")[1]
	date = strings.Split(lines[2], ":")[1]
	tags = strings.Split(strings.Split(lines[3], ":")[1], ",")
	return
}
