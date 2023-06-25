package quill

import (
	"context"
	"fmt"
	"runtime"
	"runtime/trace"
	"sync"
	"time"
)

type DataSource[T any] struct {
	commandsToSchedule chan Command
	data               T
	wg                 *sync.WaitGroup
}

func NewDataSource[T any](data T) *DataSource[T] {
	c := make(chan Command, 10)
	wg := &sync.WaitGroup{}
	go dataSourceScheduler(data, wg, c)
	return &DataSource[T]{
		data:               data,
		commandsToSchedule: c,
		wg:                 wg,
	}
}

type dataSourceWorkerJob struct {
	command     Command
	commandData any
	permissions map[string]PermissionType
}

func dataSourceWorker(index int, permissionTable *PermissionTable, wg *sync.WaitGroup, sourceData any, jobs <-chan dataSourceWorkerJob) {
	ctx, task := trace.NewTask(context.Background(), fmt.Sprintf("datasourceWorker-%d", index))
	for job := range jobs {
		PopulateView(sourceData, job.commandData)
		trace.WithRegion(ctx, "command", func() { job.command.Run() })
		permissionTable.Clear(job.permissions)
		wg.Done()
	}
	task.End()
}

func dataSourceScheduler(data any, wg *sync.WaitGroup, commands <-chan Command) {
	permissionTable := NewPermissionTable()

	numWorkers := runtime.NumCPU()
	// numWorkers = 2

	jobs := make(chan dataSourceWorkerJob, 1000)
	for i := 0; i < numWorkers; i++ {
		go dataSourceWorker(i, permissionTable, wg, data, jobs)
	}

	for command := range commands {
		commandData := command.data()
		commandsPermission := calculatePermissions(data, commandData)
		for {
			successful := permissionTable.TryAdd(commandsPermission)
			if successful {
				break
			}

			// TODO: Use a channel to wait for the permissionTable itself to
			// have changed instead of a stupid sleep (maybe take advantage of WaitGroup?)
			time.Sleep(time.Millisecond)
		}
		jobs <- dataSourceWorkerJob{
			command:     command,
			permissions: commandsPermission,
			commandData: commandData,
		}
	}
	close(jobs)
}

func (ds *DataSource[T]) RunSequentially(commands ...Command) {
	for _, c := range commands {
		commandData := c.data()
		PopulateView(ds.data, commandData)
		c.Run()
	}
}

func (ds *DataSource[T]) Run(commands ...Command) {
	ds.wg.Add(len(commands))
	for _, c := range commands {
		ds.commandsToSchedule <- c
	}
}

func (ds *DataSource[T]) Wait() {
	ds.wg.Wait()
}

func (ds *DataSource[T]) Close() {
	ds.Wait()
	close(ds.commandsToSchedule)
}
