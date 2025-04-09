package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type JobResponse struct {
    ID         string    `json:"id"`
    Command    string    `json:"command"`
    Type       string    `json:"type"`
    Status     string    `json:"status"`
    Schedule   string    `json:"schedule,omitempty"`
    CreatedAt  time.Time `json:"created_at"`
    StartedAt  time.Time `json:"started_at,omitempty"`
    FinishedAt time.Time `json:"finished_at,omitempty"`
    Output     string    `json:"output,omitempty"`
    Error      string    `json:"error,omitempty"`
}

func (s *Server) listJobs(c *gin.Context) {
    jobs := s.jobService.GetAllJobs()
    
    response := make([]JobResponse, 0, len(jobs))
    for _, job := range jobs {
        response = append(response, JobResponse{
            ID:         job.ID,
            Command:    job.Command,
            Type:       string(job.Type),
            Status:     string(job.Status),
            Schedule:   job.Schedule,
            CreatedAt:  job.CreatedAt,
            StartedAt:  job.StartedAt,
            FinishedAt: job.FinishedAt,
            Output:     job.Output,
            Error:      job.Error,
        })
    }
    
    c.JSON(http.StatusOK, response)
}

func (s *Server) getJob(c *gin.Context) {
    id := c.Param("id")
    job, err := s.jobService.GetJob(id)
    if err != nil {
        c.JSON(http.StatusNotFound, gin.H{"error": "job not found"})
        return
    }
    
    c.JSON(http.StatusOK, JobResponse{
        ID:         job.ID,
        Command:    job.Command,
        Type:       string(job.Type),
        Status:     string(job.Status),
        Schedule:   job.Schedule,
        CreatedAt:  job.CreatedAt,
        StartedAt:  job.StartedAt,
        FinishedAt: job.FinishedAt,
        Output:     job.Output,
        Error:      job.Error,
    })
}

type Server struct {
	jobService *JobService
	port       int
}

func NewServer(jobService *JobService, port int) *Server {
	return &Server{
		jobService: jobService,
		port:       port,
	}
}

func (s *Server) Start() error {
	router := gin.Default()
	
	router.Static("/static", "./static")
	
	router.LoadHTMLFiles("static/index.html")
	router.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html", nil)
	})
	
	api := router.Group("/api")
	{
		api.POST("/jobs", s.createJob)
		api.GET("/jobs", s.listJobs)
		api.GET("/jobs/:id", s.getJob)
	}
	
	return router.Run(fmt.Sprintf(":%d", s.port))
}

func (s *Server) createJob(c *gin.Context) {
	var request struct {
		Command  string `json:"command" binding:"required"`
		Type     string `json:"type" binding:"required"`
		Schedule string `json:"schedule"`
	}
	
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	var jobType JobType
	switch request.Type {
	case string(JobTypeOnce):
		jobType = JobTypeOnce
	case string(JobTypeCron):
		jobType = JobTypeCron
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid job type, must be 'once' or 'cron'"})
		return
	}
	
	job, err := s.jobService.CreateJob(request.Command, jobType, request.Schedule)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	c.JSON(http.StatusCreated, job)
}
