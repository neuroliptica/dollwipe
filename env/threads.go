// threads.go: working with 2ch threads api.
// Get all posts, get all threads on board, get random threads, etc.

package env

import (
	"dollwipe/network"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"time"
)

type Catalog struct {
	Threads []struct{ Num string }
}

type Comment struct {
	Comment string
	Num     uint64
}

type Thread struct {
	Posts []Comment
}

func getAllThreads(board string) (*Catalog, error) {
	link := fmt.Sprintf("https://2ch.hk/%s/catalog.json", board)
	cont, err := network.SendGet(link)
	if err != nil {
		return nil, err
	}
	var catalog Catalog
	json.Unmarshal(cont, &catalog)
	if len(catalog.Threads) == 0 {
		return nil, fmt.Errorf("0 тредов было найдено на доске.")
	}
	return &catalog, nil
}

func GetRandomThread(board string) (string, error) {
	catalog, err := getAllThreads(board)
	if err != nil {
		return "", err
	}
	rand.Seed(time.Now().UnixNano())
	thread := catalog.Threads[rand.Intn(len(catalog.Threads))]
	return thread.Num, nil
}

func getAllPosts(board, thread string) (*Thread, error) {
	link := fmt.Sprintf("https://2ch.hk/%s/res/%s.json", board, thread)
	cont, err := network.SendGet(link)
	if err != nil {
		return nil, err
	}
	var posts struct{ Threads []Thread }
	json.Unmarshal(cont, &posts)
	if len(posts.Threads) == 0 || len(posts.Threads[0].Posts) == 0 {
		return nil, fmt.Errorf("%s/%s не удалось получить посты.",
			board, thread)
	}
	return &(posts.Threads[0]), nil
}

func GetRandomPost(board, thread string) (*Comment, error) {
	posts, err := getAllPosts(board, thread)
	if err != nil {
		return nil, err
	}
	rand.Seed(time.Now().UnixNano())
	post := posts.Posts[rand.Intn(len(posts.Posts))]
	return &post, nil
}

// Will get all posts from all threads on board, in parallel.
// Content is about to be with the html tags, need replace.
// TODO: replace the HTML tags in content with the makaba tags.
func getPostsTexts(board string) ([]string, error) {
	catalog, err := getAllThreads(board)
	if err != nil {
		return nil, err
	}
	var (
		ch     = make(chan *Thread)
		posts  = make([]string, 0)
		failed = 0
	)
	for _, thread := range catalog.Threads {
		go func(id string) {
			t, err := getAllPosts(board, id)
			if err != nil {
				ch <- nil
			}
			ch <- t
		}(thread.Num)
	}
	for range catalog.Threads {
		t := <-ch
		if t == nil {
			failed++
			continue
		}
		for _, comment := range t.Posts {
			if comment.Comment != "" {
				posts = append(posts, comment.Comment)
			}
		}

	}
	if len(posts) == 0 {
		return nil, fmt.Errorf("не получилось найти ни одного поста.")
	}
	log.Printf("%d/%d тредов обработано; %d постов получено.",
		len(catalog.Threads)-failed, len(catalog.Threads), len(posts))
	return posts, nil
}
