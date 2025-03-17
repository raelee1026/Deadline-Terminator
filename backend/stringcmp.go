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
	filterPrefix := os.Getenv("FILTER_PREFIX")
	CourseNames = []string{}
	processedSubjects := make(map[string]bool)

	// courses, err := loadCourses("course/crawl" + filterPrefix + ".json")
	// docker
	courses, err := loadCourses("/app/course/crawl" + filterPrefix + ".json")
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
		//fmt.Printf("subject : %s here \n", subject)
		if processedSubjects[subject] {
			continue
		}

		processedSubjects[subject] = true
		
		if strings.HasPrefix(subject, filterPrefix+".") {
			re := regexp.MustCompile("^" + filterPrefix + `\.(\d+)(:|\.)`)
			match := re.FindStringSubmatch(subject)

			fmt.Printf("match : %s here \n", match)

			if len(match) > 1 {
				number := match[1]
				found := false
				for _, course := range courses {
					uniqueIDLast6 := course.UniqueID[len(course.UniqueID)-6:]
					if uniqueIDLast6 == number {
						CourseNames = append(CourseNames, course.CosCname)
						found = true
						break
					}
				}
				if !found {
					CourseNames = append(CourseNames, "")
				}
			}
		}
	}

	for i := 0; i < len(CourseNames); i++ {
    fmt.Printf("Course %d: '%s'\n", i+1, CourseNames[i])
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
