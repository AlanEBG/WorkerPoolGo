package main

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// Estructura que representa una tarea
type Job struct {
	ID    int
	Value int
}

// Estructura con el resultado del procesamiento
type Result struct {
	Job       Job
	WorkerID  int
	Output    int
	Timestamp time.Time
}

// Función que ejecuta cada worker (trabajador)
func worker(id int, ctx context.Context, jobs <-chan Job, results chan<- Result, wg *sync.WaitGroup) {
	defer wg.Done()

	for {
		select {
		case <-ctx.Done():
			fmt.Printf("Worker %d: cancelado por contexto\n", id)
			return
		case job, ok := <-jobs:
			if !ok {
				fmt.Printf("Worker %d: ya no hay más tareas, cerrando\n", id)
				return
			}

			fmt.Printf("Worker %d iniciando tarea %d\n", id, job.ID)

			// Simula trabajo (1 segundo)
			select {
			case <-ctx.Done():
				fmt.Printf("Worker %d: tarea %d cancelada\n", id, job.ID)
				return
			case <-time.After(time.Second):
			}

			output := job.Value * 2

			result := Result{
				Job:       job,
				WorkerID:  id,
				Output:    output,
				Timestamp: time.Now(),
			}

			select {
			case <-ctx.Done():
				return
			case results <- result:
				fmt.Printf("Worker %d terminó tarea %d (resultado: %d)\n", id, job.ID, output)
			}
		}
	}
}

// Función que recibe y muestra los resultados
func processResults(results <-chan Result, numJobs int, done chan<- bool) {
	var list []Result

	for i := 0; i < numJobs; i++ {
		res := <-results
		list = append(list, res)
	}

	fmt.Println("\n=== RESULTADOS ===")
	for _, r := range list {
		fmt.Printf("Tarea %d hecha por worker %d: %d -> %d (%s)\n",
			r.Job.ID, r.WorkerID, r.Job.Value, r.Output, r.Timestamp.Format("15:04:05.000"))
	}

	done <- true
}

func main() {
	fmt.Println("=== WORKER POOL EN GO ===\n")

	const (
		numWorkers = 4
		numJobs    = 20
	)

	start := time.Now()

	// Contexto con cancelación
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Canales
	jobs := make(chan Job, numJobs)
	results := make(chan Result, numJobs)

	var wg sync.WaitGroup

	// Lanzar los workers
	fmt.Printf("Arrancando %d workers...\n\n", numWorkers)
	for i := 1; i <= numWorkers; i++ {
		wg.Add(1)
		go worker(i, ctx, jobs, results, &wg)
	}

	// Enviar las tareas
	fmt.Printf("Enviando %d tareas...\n\n", numJobs)
	go func() {
		for i := 1; i <= numJobs; i++ {
			jobs <- Job{ID: i, Value: i * 10}
		}
		close(jobs)
	}()

	done := make(chan bool)
	go processResults(results, numJobs, done)

	// Esperar a que terminen los workers
	wg.Wait()
	close(results)

	<-done

	duration := time.Since(start)

	fmt.Println("\n=== ESTADÍSTICAS ===")
	fmt.Printf("Total de tareas: %d\n", numJobs)
	fmt.Printf("Workers activos: %d\n", numWorkers)
	fmt.Printf("Duración total: %s\n", duration)
	fmt.Printf("Promedio por tarea: %s\n", duration/time.Duration(numJobs))

	// Comparar con ejecución secuencial
	seqTime := time.Second * time.Duration(numJobs)
	saved := seqTime - duration
	fmt.Printf("Tiempo estimado (secuencial): %s\n", seqTime)
	fmt.Printf("Ahorro de tiempo: %s (%.2f%% más rápido)\n",
		saved, float64(saved)/float64(seqTime)*100)

	fmt.Println("\nProceso completado ✓")
}
