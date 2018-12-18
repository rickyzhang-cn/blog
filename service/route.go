package service

import (
	"html/template"
	"io/ioutil"
	"net/http"

	"encoding/json"

	"github.com/gin-gonic/gin"
	"github.com/russross/blackfriday"
)

const (
	CATEGORYMAP = "CategoryMap"
	TAGMAP      = "TagMap"
	POSTS       = "Posts"
	CATEGORIES  = "Categories"
	TAGS        = "Tags"
)

type Index struct {
	Posts []*Post `json:"posts"`
}

type Post struct {
	Title      string   `json:"title"`
	FileName   string   `json:"file_name"`
	Date       string   `json:"date"`
	Categories []string `json:"categories"`
	Tags       []string `json:"tags"`
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
		c.Set(POSTS, info.Posts)
		categoryM := make(map[string][]*Post, 16)
		tagM := make(map[string][]*Post, 16)
		for _, post := range info.Posts {
			cates, tags := post.Categories, post.Tags
			for _, cate := range cates {
				categoryM[cate] = append(categoryM[cate], post)
			}
			for _, tag := range tags {
				tagM[tag] = append(tagM[tag], post)
			}
		}
		c.Set(CATEGORYMAP, categoryM)
		c.Set(TAGMAP, tagM)
		categories := make([]string, 0, 16)
		tags := make([]string, 0, 16)
		for cate := range categoryM {
			categories = append(categories, cate)
		}
		for tag := range tagM {
			tags = append(tags, tag)
		}
		c.Set(CATEGORIES, categories)
		c.Set(TAGS, tags)
		c.Next()
	}
}

func Home(c *gin.Context) {
	info, exist := c.Get(POSTS)
	if !exist {
		c.Abort()
	}
	posts, ok := info.([]*Post)
	if !ok {
		c.Abort()
	}
	categories := c.GetStringSlice(CATEGORIES)
	tags := c.GetStringSlice(TAGS)
	c.HTML(http.StatusOK, "home.html", gin.H{
		POSTS:      posts,
		CATEGORIES: categories,
		TAGS:       tags,
	})
}

func About(c *gin.Context) {
	categories := c.GetStringSlice(CATEGORIES)
	tags := c.GetStringSlice(TAGS)
	path := "./page/about.md"
	content, err := ioutil.ReadFile(path)
	if err != nil {
		c.Abort()
	}
	html := blackfriday.MarkdownCommon(content)
	c.HTML(http.StatusOK, "page.html", gin.H{
		CATEGORIES: categories,
		TAGS:       tags,
		"Content":  template.HTML(html),
	})
}

func GetCategory(c *gin.Context) {
	categories := c.GetStringSlice(CATEGORIES)
	tags := c.GetStringSlice(TAGS)
	cate := c.Param("name")
	info, ok := c.Get(CATEGORYMAP)
	if !ok {
		c.Abort()
	}
	cateM, ok := info.(map[string][]*Post)
	if !ok {
		c.Abort()
	}
	posts := cateM[cate]
	if !ok {
		c.Abort()
	}
	c.HTML(http.StatusOK, "home.html", gin.H{
		POSTS:      posts,
		CATEGORIES: categories,
		TAGS:       tags,
	})

}

func GetTag(c *gin.Context) {
	categories := c.GetStringSlice(CATEGORIES)
	tags := c.GetStringSlice(TAGS)
	tag := c.Param("name")
	info, ok := c.Get(TAGMAP)
	if !ok {
		c.Abort()
	}
	tagM, ok := info.(map[string][]*Post)
	if !ok {
		c.Abort()
	}
	posts := tagM[tag]
	c.HTML(http.StatusOK, "home.html", gin.H{
		POSTS:      posts,
		CATEGORIES: categories,
		TAGS:       tags,
	})
}

func GetPost(c *gin.Context) {
	categories := c.GetStringSlice(CATEGORIES)
	tags := c.GetStringSlice(TAGS)
	path := "./post/" + c.Param("name")
	content, err := ioutil.ReadFile(path)
	if err != nil {
		c.Abort()
	}
	info, exist := c.Get(POSTS)
	if !exist {
		c.Abort()
	}
	posts, ok := info.([]*Post)
	if !ok {
		c.Abort()
	}
	var title, date string
	for _, p := range posts {
		if p.FileName == c.Param("name") {
			title, date = p.Title, p.Date
		}
	}
	html := blackfriday.MarkdownCommon(content)
	c.HTML(http.StatusOK, "post.html", gin.H{
		CATEGORIES: categories,
		TAGS:       tags,
		"Title":    title,
		"Date":     date,
		"Content":  template.HTML(html),
	})
}
