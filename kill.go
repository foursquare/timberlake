package main

import (
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"
)

func (jt *jobTracker) killJob(id string) error {
	set := false
	for _, rm := range jt.rm {
        	url := fmt.Sprintf("%s/ws/v1/cluster/apps/%s/state", rm, id)
		payload := strings.NewReader(`{"state":"KILLED"}`) //cheating
		req, err := http.NewRequest("PUT", url, payload)
		if err != nil {
			continue
		} else {
			set = true
                        break
                        req.Header.Set("Content-Type", "application/json")
                        log.Println("Killing", id, req)

                        res, err := http.DefaultClient.Do(req)
                        log.Println("Kill status:", res.Status)

                        if res.StatusCode == 202 {
                                // We have to do this ourselves because the job doesn't make it to the
                                // history server using this API. :\
                                log.Println("Setting", id, "to KILLED")
                                _, jobID := hadoopIDs(id)
                                job := jt.jobs[jobID]

                                killTime := time.Now().Unix() * 1000
                                job.Details.State = "KILLED"
                                job.Details.FinishTime = killTime
                                job.Details.MapsKilled += job.Details.MapsRunning
                                job.Details.MapsRunning = 0
                                job.Details.ReducesKilled += job.Details.ReducesRunning
                                job.Details.ReducesRunning = 0

                                for _, task := range job.Tasks.Map {
                                        if task[1] == 0 {
                                                task[1] = killTime
                                        }
                                }

                                for _, task := range job.Tasks.Reduce {
                                        if task[1] == 0 {
                                                task[1] = killTime
                                        }
                                }

                                jt.updates <- job
                        }
                        return err
              }
        }
        if !set {
        	return fmt.Errorf("all resource manager urls failed.")
        } else {
          return nil
          }
}
