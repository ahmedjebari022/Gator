package main

import (
	"context"
	"encoding/xml"
	"fmt"
	"html"
	"io"
	"net/http"
)

type RSSFeed struct {
	Channel struct {
		Title       string    `xml:"title"`
		Link        string    `xml:"link"`
		Description string    `xml:"description"`
		Item        []RSSItem `xml:"item"`
	} `xml:"channel"`
}

type RSSItem struct {
	Title       string `xml:"title"`
	Link        string `xml:"link"`
	Description string `xml:"description"`
	PubDate     string `xml:"pubDate"`
}


func FetchFeed(ctx context.Context, feedURL string)(*RSSFeed, error){
	req, err := http.NewRequestWithContext(ctx,"GET",feedURL,nil)
	if err != nil {
		return &RSSFeed{},err
	}
	req.Header.Set("User-Agent","gator")
	client := http.DefaultClient
	res, err := client.Do(req)
	if err != nil {
		return &RSSFeed{},err
	}
	if res.StatusCode != 200{
		return nil,fmt.Errorf("erros status code: %d",res.StatusCode)
	}
	defer res.Body.Close()
	data, err := io.ReadAll(res.Body)
	if err != nil {
		return &RSSFeed{},err
	}
	var rssFeed RSSFeed
	err = xml.Unmarshal(data,&rssFeed)
	if err != nil {
		return &RSSFeed{},err
	}
	rssFeed.Channel.Description = html.UnescapeString(rssFeed.Channel.Description)
	rssFeed.Channel.Title = html.UnescapeString(rssFeed.Channel.Title)
	for i, v := range rssFeed.Channel.Item{
		rssFeed.Channel.Item[i].Description = html.UnescapeString(v.Description)
		rssFeed.Channel.Item[i].Title = html.UnescapeString(v.Title)
	}
	return &rssFeed,nil
}