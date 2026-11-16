// Package di binds application ports to adapters and owns their lifetime.
package di

import (
	"io"
	"sync"

	"pos/internal/application"
	"pos/internal/cli"
	"pos/internal/domain"
	"pos/internal/infrastructure/csvexport"
	"pos/internal/infrastructure/memory"
)

// Container is a typed DI container. Providers create one instance per container
// on first use. Dependencies are explicit; no global registry or reflection is used.
// The container is accessed only at the composition root, never by domain or use cases.
type Container struct {
	students func() application.StudentRepository
	rubrics  func() application.RubricRepository
	forms    func() application.EvaluationFormRepository
	exporter func() application.ResultExporter

	studentService func() *application.StudentService
	rubricService  func() *application.RubricService
	formService    func() *application.EvaluationFormService
	resultService  func() *application.ResultService
	facade         func() *application.Facade
}

func New() *Container {
	c := &Container{}
	c.students = sync.OnceValue(func() application.StudentRepository {
		return memory.NewStudentRepository()
	})
	c.rubrics = sync.OnceValue(func() application.RubricRepository {
		return memory.NewRubricRepository()
	})
	c.forms = sync.OnceValue(func() application.EvaluationFormRepository {
		return memory.NewEvaluationFormRepository()
	})
	c.exporter = sync.OnceValue(func() application.ResultExporter {
		return csvexport.NewResultExporter()
	})
	c.studentService = sync.OnceValue(func() *application.StudentService {
		return application.NewStudentService(c.students())
	})
	c.rubricService = sync.OnceValue(func() *application.RubricService {
		return application.NewRubricService(c.rubrics())
	})
	c.formService = sync.OnceValue(func() *application.EvaluationFormService {
		return application.NewEvaluationFormService(c.students(), c.rubrics(), c.forms())
	})
	c.resultService = sync.OnceValue(func() *application.ResultService {
		return application.NewResultService(c.students(), c.forms(), domain.NewEvaluationService(domain.HundredthsScorePolicy{}), c.exporter())
	})
	c.facade = sync.OnceValue(func() *application.Facade {
		return application.NewFacade(c.StudentService(), c.RubricService(), c.EvaluationFormService(), c.ResultService())
	})
	return c
}

func (c *Container) StudentService() *application.StudentService { return c.studentService() }

func (c *Container) RubricService() *application.RubricService { return c.rubricService() }

func (c *Container) EvaluationFormService() *application.EvaluationFormService {
	return c.formService()
}

func (c *Container) ResultService() *application.ResultService { return c.resultService() }

func (c *Container) Facade() *application.Facade { return c.facade() }

// CLI creates a session with its own input/output and the container's shared services.
func (c *Container) CLI(input io.Reader, output io.Writer) *cli.App {
	return cli.New(input, output, c.Facade())
}
