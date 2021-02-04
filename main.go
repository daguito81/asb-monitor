package main

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	servicebus "github.com/Azure/azure-service-bus-go"
	tm "github.com/buger/goterm"
	"github.com/daguito81/sbmgmt"
	"github.com/fatih/color"
	"github.com/olekukonko/tablewriter"
	log "github.com/sirupsen/logrus"
)

var sb *servicebus.Namespace
var red = color.New(color.FgRed).SprintFunc()
var yellow = color.New(color.FgYellow).SprintFunc()
var green = color.New(color.FgHiGreen).SprintFunc()

var SBName string

func init(){
	if len(os.Args) != 2{
		log.Error("Need to provide 1 argument: SBName (from .env file)")
		time.Sleep(5*time.Second)
		os.Exit(1)
	}
	sbName := os.Args[1]
	var err error
	sb, err = sbmgmt.GetServiceBusNamespace(sbName)
	if err != nil {
		log.Fatal("Can't get a connection to ServiceBus")
		os.Exit(1)
	}
}


func main(){
	tm.Clear()
	for {
		data := getData()
		tm.MoveCursor(1, 1)
		buildScreen(data)
		time.Sleep(5*time.Second)
		tm.Flush()
	}


}

func getData() [][]string {
	topics, err := getTopics()
	if err != nil {
		log.Fatal("Problem getting Topics: ", err)
	}
	var data [][]string

	for _, topic := range topics {
		log.Debug("Getting subs for topic: ", topic.Name)
		subs, err := getSubs(topic.Name)
		if err != nil {
			log.Error("Error getting Subs for: ", topic.Name, " : ", err)
		}
		for _, sub := range subs {
			topicName := topic.Name
			var subName string
			if sub.Name == "azurefunction1"{
				subName = red(sub.Name)
			} else {
				subName = sub.Name
			}
			subCount := *sub.MessageCount
			var subCountResult string
			if subCount > 1000 {
				subCountResult = red(strconv.FormatInt(subCount, 10))
			} else if subCount > 100 {
				subCountResult = yellow(strconv.FormatInt(subCount, 10))
			} else {
				subCountResult = green(strconv.FormatInt(subCount, 10))
			}
			data = append(data, []string{topicName, subName, subCountResult})

		}

	}
	return data
}

func getSubs(topicName string)([]*servicebus.SubscriptionEntity, error){
	sm, err := sb.NewSubscriptionManager(topicName)
	if err != nil {
		log.Error("Error getting sub manager:", err)
		return nil, err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	subs, err := sm.List(ctx)
	if err != nil{
		log.Error("Error getting subs: ", err)
		return nil, err
	}
	return subs, nil
}


func getTopics() ([]*servicebus.TopicEntity, error){
	top := sb.NewTopicManager()
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	topics, err := top.List(ctx)
	if err != nil {
		log.Fatal("Can't get list of topics")
		return nil, err
	}
	return topics, nil
}


func buildScreen(data [][]string) {
	printAll := ""
	i := utf8.RuneCountInString(sb.Name)
	maxSize := 46
	var name string
	if i%2 == 0 {
		name = green(sb.Name)
	} else {
		name = red(sb.Name + " ")
	}
	spacer := strings.Repeat(" ", (maxSize-6-i)/2)
	printAll += fmt.Sprint(strings.Repeat("#", maxSize) + "\n")
	printAll += fmt.Sprintf("###%v%s%v###\n", spacer, name, spacer)
	printAll += fmt.Sprint(strings.Repeat("#", maxSize) + "\n")
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"Topic", "Sub", "Count"})
	table.SetAutoMergeCells(true)
	for _, v := range data{
		table.Append(v)
	}
	fmt.Print(printAll)
	table.Render()

}