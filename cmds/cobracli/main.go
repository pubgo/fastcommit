package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type Task struct {
	ID          int        `json:"id"`
	Description string     `json:"description"`
	Priority    string     `json:"priority"`
	Completed   bool       `json:"completed"`
	CreatedAt   time.Time  `json:"created_at"`
	CompletedAt *time.Time `json:"completed_at,omitempty"`
}

type TaskManager struct {
	Tasks    []Task `json:"tasks"`
	NextID   int    `json:"next_id"`
	FilePath string `json:"-"`
}

var (
	taskFile    string
	priority    string
	showAll     bool
	taskManager *TaskManager
)

var rootCmd = &cobra.Command{
	Use:   "taskman",
	Short: "A personal task manager",
	Long: `taskman is a CLI task manager that helps you organize your work.

Store tasks locally with priorities, mark them complete, and keep
track of your productivity over time.`,
	PersistentPreRun:  loadTasks,
	PersistentPostRun: saveTasks,
}

var addCmd = &cobra.Command{
	Use:   "add [description]",
	Short: "Add a new task",
	Long:  "Add a new task with description and optional priority (high, medium, low)",
	Args:  cobra.MinimumNArgs(1),
	Run:   addTask,
}

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all tasks",
	Long:  "List tasks with optional filtering by completion status",
	Run:   listTasks,
}

var completeCmd = &cobra.Command{
	Use:   "complete [task-id]",
	Short: "Mark a task as completed",
	Args:  cobra.ExactArgs(1),
	Run:   completeTask,
}

var deleteCmd = &cobra.Command{
	Use:   "delete [task-id]",
	Short: "Delete a task",
	Args:  cobra.ExactArgs(1),
	Run:   deleteTask,
}

func init() {
	// Persistent flags available to all commands
	rootCmd.PersistentFlags().StringVar(&taskFile, "file", "", "task file (default is $HOME/.taskman.json)")

	// Local flags for specific commands
	addCmd.Flags().StringVarP(&priority, "priority", "p", "medium", "task priority (high, medium, low)")
	listCmd.Flags().BoolVarP(&showAll, "all", "a", false, "show completed tasks too")

	// Add subcommands
	rootCmd.AddCommand(addCmd, listCmd, completeCmd, deleteCmd)

	// Setup configuration
	setupConfig()
}

func setupConfig() {
	viper.SetConfigName("taskman")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("$HOME")
	viper.SetDefault("priority", "medium")
	viper.SetDefault("file", filepath.Join(os.Getenv("HOME"), ".taskman.json"))

	viper.ReadInConfig()
}

func getTaskFile() string {
	if taskFile != "" {
		return taskFile
	}
	return viper.GetString("file")
}

func loadTasks(cmd *cobra.Command, args []string) {
	file := getTaskFile()
	taskManager = &TaskManager{
		Tasks:    make([]Task, 0),
		NextID:   1,
		FilePath: file,
	}

	if data, err := os.ReadFile(file); err == nil {
		json.Unmarshal(data, taskManager)
	}
}

func saveTasks(cmd *cobra.Command, args []string) {
	data, err := json.MarshalIndent(taskManager, "", "  ")
	if err != nil {
		fmt.Printf("Error saving tasks: %v\n", err)
		return
	}

	os.WriteFile(taskManager.FilePath, data, 0644)
}

func addTask(cmd *cobra.Command, args []string) {
	description := strings.Join(args, " ")

	// Validate priority
	validPriorities := map[string]bool{"high": true, "medium": true, "low": true}
	if !validPriorities[priority] {
		fmt.Printf("Invalid priority '%s'. Use: high, medium, or low\n", priority)
		os.Exit(1)
	}

	task := Task{
		ID:          taskManager.NextID,
		Description: description,
		Priority:    priority,
		Completed:   false,
		CreatedAt:   time.Now(),
	}

	taskManager.Tasks = append(taskManager.Tasks, task)
	taskManager.NextID++

	fmt.Printf("Added task #%d: %s [%s]\n", task.ID, task.Description, task.Priority)
}

func listTasks(cmd *cobra.Command, args []string) {
	if len(taskManager.Tasks) == 0 {
		fmt.Println("No tasks found.")
		return
	}

	fmt.Printf("%-4s %-10s %-50s %-10s %s\n", "ID", "STATUS", "DESCRIPTION", "PRIORITY", "CREATED")
	fmt.Println(strings.Repeat("-", 90))

	for _, task := range taskManager.Tasks {
		if !showAll && task.Completed {
			continue
		}

		status := "PENDING"
		if task.Completed {
			status = "DONE"
		}

		fmt.Printf("%-4d %-10s %-50s %-10s %s\n",
			task.ID,
			status,
			truncate(task.Description, 50),
			strings.ToUpper(task.Priority),
			task.CreatedAt.Format("2006-01-02"))
	}
}

func completeTask(cmd *cobra.Command, args []string) {
	id, err := strconv.Atoi(args[0])
	if err != nil {
		fmt.Printf("Invalid task ID: %s\n", args[0])
		os.Exit(1)
	}

	for i := range taskManager.Tasks {
		if taskManager.Tasks[i].ID == id {
			if taskManager.Tasks[i].Completed {
				fmt.Printf("Task #%d is already completed\n", id)
				return
			}

			now := time.Now()
			taskManager.Tasks[i].Completed = true
			taskManager.Tasks[i].CompletedAt = &now

			fmt.Printf("Completed task #%d: %s\n", id, taskManager.Tasks[i].Description)
			return
		}
	}

	fmt.Printf("Task #%d not found\n", id)
	os.Exit(1)
}

func deleteTask(cmd *cobra.Command, args []string) {
	id, err := strconv.Atoi(args[0])
	if err != nil {
		fmt.Printf("Invalid task ID: %s\n", args[0])
		os.Exit(1)
	}

	for i, task := range taskManager.Tasks {
		if task.ID == id {
			// Remove task from slice
			taskManager.Tasks = append(taskManager.Tasks[:i], taskManager.Tasks[i+1:]...)
			fmt.Printf("Deleted task #%d: %s\n", id, task.Description)
			return
		}
	}

	fmt.Printf("Task #%d not found\n", id)
	os.Exit(1)
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max-3] + "..."
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
