package main

import "testing"

func TestGetHash(t *testing.T) {
	hash, _ := getHash("https://red-panda.me/img/botError.png")
	if hash != "h9FxvZ2mhQucUimxYcohhf3s79U=" {
		t.Error("Hash was not correct, hash was " + hash)
	}
}

func TestImageDecodingCorrent(t *testing.T) {
	result := testImg("https://red-panda.me/img/botError.png")
	if !result {
		t.Error("testImg failed when correct image was inputted")
	}
}

func TestImageDecodingIncorrent(t *testing.T) {
	result := testImg("https://red-panda.me/")
	if result {
		t.Error("testImg passed when it should of failed")
	}
}

func TestGetImgWithImg(t *testing.T) {
	mockBing := bingAnswer{
		Value: []bingValue{
			{EncodingFormat: "png", ContentURL: "https://red-panda.me/img/connor10.9.png"},
		},
	}
	URL := getImage(mockBing)
	if URL != "https://red-panda.me/img/connor10.9.png" {
		t.Errorf("Test image failed, URL was " + URL)
	}
}

func TestGetImgWithNo(t *testing.T) {
	mockBing := bingAnswer{}
	URL := getImage(mockBing)
	if URL != "https://red-panda.me/img/botError.png" {
		t.Errorf("Test image failed, URL was " + URL)
	}
}
