package main

import (
    "fmt"
    "github.com/trietmn/go-wiki"
    gt "github.com/bas24/googletranslatefree"
    "github.com/redis/go-redis/v9"
    "context"
    "net/http"
    "github.com/gin-gonic/gin"
    "database/sql"
    _ "github.com/lib/pq"
)

func main() {
    // Search for the Wikipedia page title
    // search_result, _, err := gowiki.Search("who is batman", 3, false)
    // if err != nil {
    //     fmt.Println(err)
    // }
    // fmt.Printf("This is your search result: %v\n", search_result[0])
    
    router := gin.Default()
    ctx := context.Background()
    
	// Connect to Redis
	rdb := redis.NewClient(&redis.Options{
		Addr: "localhost:6379", // Redis server address
		DB:   0,                // Default DB
	})
	
	// test the connection 
	_, err := rdb.Ping(ctx).Result()
	if err != nil {
		fmt.Println("Could not connect to Redis: %v", err)
	}
    fmt.Println("Connected to Redis!")
   
    // title := "superMan" 
    fmt.Println("localhost:3000")
    router.GET("/", func(c *gin.Context) {
        title := c.Query("title")
        
        val, err := rdb.Get(ctx, title).Result()
        if err == nil {
            
            c.JSON(http.StatusOK, gin.H{
                "title": title,
                "content": val,
            })
            
            return
        }
        
        connStr := ""
        // Open a connection
	    db, _ := sql.Open("postgres", connStr)
	
        defer db.Close()
        
        // Get the page
        page, err := gowiki.GetPage(title, -1, false, true)
        if err != nil {
            c.JSON(404, gin.H{
                "messge": err,
            })
        }
        
        // Get the content of the page
        summrary, err := page.GetSummary()
        if err != nil {
            c.JSON(404, gin.H{
                "messge": err,
            })
        }
        // fmt.Printf("This is the page content: %v\n", summrary)
        
        result, _ := gt.Translate(summrary, "en", "ckb")
        // fmt.Println(result)
        
        
        err = rdb.Set(ctx, title, result, 0).Err()
        if err != nil {
            c.JSON(404, gin.H{
                "messge": err,
            })
        }
        
        sql := fmt.Sprintf(`INSERT INTO wiki (content, title) VALUES ('%v', '%v');`, result, title )
        
        db.Query(sql)
        
        c.JSON(http.StatusOK, gin.H{
            "title": title,
            "content": result,
        })
        
    })
    
    router.Run(":3000")
}

