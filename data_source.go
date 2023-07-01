package quill

import (
	"runtime"
	"sync"
)

type DataSource[T any] struct {
	commandsToSchedule chan Command
	data               T
	wg                 *sync.WaitGroup
}

func NewDataSource[T any](data T) *DataSource[T] {
	return NewDataSourceWithPoolSize[T](data, runtime.NumCPU())
}

func NewDataSourceWithPoolSize[T any](data T, pool int) *DataSource[T] {
	c := make(chan Command, 10)
	wg := &sync.WaitGroup{}
	go dataSourceScheduler(data, wg, c, pool)
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

func dataSourceWorker(
	index int,
	permissionTable *PermissionTable,
	wg *sync.WaitGroup,
	sourceData any,
	jobs <-chan *dataSourceWorkerJob,
) {
	// ctx, task := trace.NewTask(context.Background(), fmt.Sprintf("datasourceWorker-%d", index))
	for job := range jobs {
		applyChanges := PopulateView(sourceData, job.commandData)
		// trace.WithRegion(ctx, "command", func() { job.command.Run() })
		job.command.Run()
		applyChanges.Apply()
		permissionTable.Clear(job.permissions)
		wg.Done()
	}
	// task.End()
}

func dataSourceScheduler(data any, wg *sync.WaitGroup, commands <-chan Command, poolSize int) {
	permissionTable := NewPermissionTable()

	numWorkers := poolSize
	if numWorkers > 1 {
		numWorkers -= 1 // Leave one cpu unallocated for the scheduler goroutine
	}
	// numWorkers = 2

	jobs := make(chan *dataSourceWorkerJob, 1000)
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

			// TODO: Be smarter about waiting until it's valid to try to add again.
			version := permissionTable.Version()
			newVersion := version
			for version == newVersion {
				newVersion = permissionTable.Version()
			}
		}

		jobs <- &dataSourceWorkerJob{
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
