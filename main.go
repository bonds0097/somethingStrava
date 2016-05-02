package main

import (
	"fmt"
	"log"
	"os"
	"sort"
	"strconv"
	"time"

	"github.com/strava/go.strava"
)

var (
	stravaToken string
)

const (
	metersToMiles = 0.000621371
	metersToFeet  = 3.28084
	speedToPace   = 26.8224 * 60
	mpsTomph      = 2.23694
)

func init() {
	stravaToken = os.Getenv("STRAVA_TOKEN")
	if stravaToken == "" {
		log.Fatalf("Failed to retrieve Strava token.")
	}
}

type Week struct {
	Running       []string
	RunningTotal  float64
	Cycling       []string
	CyclingTotal  float64
	Swimming      []string
	SwimmingTotal float64
}

func main() {
	var week Week

	client := strava.NewClient(stravaToken)
	service := strava.NewCurrentAthleteService(client)

	athlete, err := service.Get().Do()
	if err != nil {
		log.Fatalf("Failed to retrieve athlete: %s", err)
	}
	stats, err := strava.NewAthletesService(client).Stats(athlete.Id).Do()
	if err != nil {
		log.Fatalf("Failed to retrieve athlete stats: %s", err)
	}

	cyclingYTD := stats.YTDRideTotals.Distance * metersToMiles

	today := time.Now().Round(time.Hour * 24)
	lastSunday := today.AddDate(0, 0, int(time.Sunday-today.Weekday()))
	thisMonday := int(today.AddDate(0, 0, int(time.Monday-today.Weekday())).Unix())
	lastWeek := int(lastSunday.AddDate(0, 0, -7).Unix())
	activities, err := service.ListActivities().Before(thisMonday).After(lastWeek).Do()
	if err != nil {
		log.Fatalf("Failed to retrieve activities: %s", err)
	}

	for _, activity := range activities {
		url := "https://www.strava.com/activities/" + strconv.Itoa(int(activity.Id))
		s := fmt.Sprintf(`[url="%s"]%s[/url]: %s, %4.2f miles, %4.0fft elev. gain`,
			url,
			activity.StartDate.Format("01/02/2006"),
			time.Duration(activity.ElapsedTime)*time.Second,
			activity.Distance*metersToMiles,
			activity.TotalElevationGain*metersToFeet,
		)

		switch activity.Type {
		case "Run":
			s += fmt.Sprintf(", %s per mile",
				time.Duration((1/activity.AverageSpeed)*speedToPace)*time.Second)
			week.Running = append(week.Running, s)
			week.RunningTotal += activity.Distance * metersToMiles
		case "Ride":
			s += fmt.Sprintf(", %2.1f MPH", activity.AverageSpeed*mpsTomph)
			week.Cycling = append(week.Cycling, s)
			week.CyclingTotal += activity.Distance * metersToMiles
		case "Swim":
			s = fmt.Sprintf(`[url="%s"]%s[/url]: %s, %4.2f meters`,
				url,
				activity.StartDate.Format("01/02/2006"),
				time.Duration(activity.ElapsedTime)*time.Second,
				activity.Distance,
			)
			week.Swimming = append(week.Swimming, s)
			week.SwimmingTotal += activity.Distance

		}

	}

	sort.Sort(sort.Reverse(sort.StringSlice(week.Running)))
	sort.Sort(sort.Reverse(sort.StringSlice(week.Cycling)))

	if len(week.Running) > 0 {
		fmt.Print("[b]Running[/b]\n\n[fixed]")
		for _, activity := range week.Running {
			fmt.Println(activity)
		}
		fmt.Printf("[/fixed]\n[b]Weekly Total:[/b] %2.2f miles\n\n", week.RunningTotal)
	}

	if len(week.Cycling) > 0 {
		fmt.Print("[b]Cycling[/b]\n\n[fixed]")
		for _, activity := range week.Cycling {
			fmt.Println(activity)
		}
		fmt.Printf("[/fixed]\n[b]Weekly Total:[/b] %2.2f miles\n", week.CyclingTotal)
		fmt.Printf("[/fixed]\n[b]YTD Total:[/b] %2.2f miles\n\n", cyclingYTD)
	}

	if len(week.Swimming) > 0 {
		fmt.Print("[b]Swimming[/b]\n\n[fixed]")
		for _, activity := range week.Swimming {
			fmt.Println(activity)
		}
		fmt.Printf("[/fixed]\n[b]Weekly Total:[/b] %2.2f meters\n\n", week.SwimmingTotal)
	}

}
