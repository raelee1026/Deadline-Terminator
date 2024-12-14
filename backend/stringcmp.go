package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"regexp"
	"strings"

	"google.golang.org/api/gmail/v1"
)

type Course struct {
	CosCode  string `json:"cos_code"`
	CosCname string `json:"cos_cname"`
	UniqueID string `json:"unique_id"`
}

func ProcessString(message []gmail.Message) ([]string, error) {
	if len(message) == 0 {
		log.Println("No messages to process.")
		return []string{}, nil
	}

	courses, err := loadCourses("../backend/course/crawl1131.json")
	if err != nil {
		return nil, fmt.Errorf("failed to load courses: %v", err)
	}

	for _, msg := range message {
		var subject string
		for _, header := range msg.Payload.Headers {
			if header.Name == "Subject" {
				subject = header.Value
				break
			}
		}
		if strings.HasPrefix(subject, "1131.") {
			re := regexp.MustCompile(`^1131\.(\d+):`)
			match := re.FindStringSubmatch(subject)
			if len(match) > 1 {
				number := match[1]

				for _, course := range courses {
					uniqueIDLast6 := course.UniqueID[len(course.UniqueID)-6:]
					if uniqueIDLast6 == number {
						CourseNames = append(CourseNames, course.CosCname)
					}
				}
			}
		}
	}

	return CourseNames, nil
}

func loadCourses(courseFile string) ([]Course, error) {
	file, err := os.Open(courseFile)
	if err != nil {
		return nil, fmt.Errorf("failed to open course file: %v", err)
	}
	defer file.Close()

	var courses []Course
	if err := json.NewDecoder(file).Decode(&courses); err != nil {
		return nil, fmt.Errorf("failed to decode course file: %v", err)
	}

	return courses, nil
}
