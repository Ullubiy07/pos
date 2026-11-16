package cli

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"pos/internal/application"

	"github.com/google/uuid"
)

type App struct {
	input       *bufio.Scanner
	output      io.Writer
	students    *application.StudentService
	rubrics     *application.RubricService
	forms       *application.EvaluationFormService
	results     *application.ResultService
	studentID   uuid.UUID
	rubricID    uuid.UUID
	studentName string
	rubricName  string
}

func New(input io.Reader, output io.Writer, students *application.StudentService, rubrics *application.RubricService, forms *application.EvaluationFormService, results *application.ResultService) *App {
	scanner := bufio.NewScanner(input)
	scanner.Buffer(make([]byte, 4096), 1024*1024)
	return &App{input: scanner, output: output, students: students, rubrics: rubrics, forms: forms, results: results}
}

func (a *App) Run(ctx context.Context) error {
	fmt.Fprintln(a.output, "ПОС — преподавательская оценка студентов")
	fmt.Fprintln(a.output, "Данные хранятся в памяти до выхода. Итог можно экспортировать в CSV.")
	for {
		if err := ctx.Err(); err != nil {
			return err
		}
		a.menu()
		command, err := a.read("Команда: ")
		if errors.Is(err, io.EOF) {
			return nil
		}
		if err != nil {
			return err
		}
		if command == "0" {
			return nil
		}
		if err := a.execute(ctx, command); err != nil {
			if errors.Is(err, io.EOF) {
				return nil
			}
			if a.input.Err() != nil {
				return err
			}
			fmt.Fprintf(a.output, "Ошибка: %v\n", err)
		}
	}
}

func (a *App) menu() {
	fmt.Fprintf(a.output, "\nСтудент: %s | Рубрикатор: %s\n", selected(a.studentName), selected(a.rubricName))
	fmt.Fprintln(a.output, `1  — Добавить студента
2  — Список студентов / выбрать студента
3  — Создать рубрикатор
4  — Список рубрикаторов / выбрать рубрикатор
5  — Показать критерии выбранного рубрикатора
6  — Создать форму для выбранного студента и рубрикатора
7  — Показать форму и оценки
8  — Выставить оценку
9  — Исправить оценку
10 — Отменить оценку
11 — Завершить форму
12 — Открыть форму для исправлений
13 — Рассчитать итог
14 — Экспортировать итог в CSV
0  — Выход`)
}

func selected(name string) string {
	if name == "" {
		return "не выбран"
	}
	return name
}

func (a *App) execute(ctx context.Context, command string) error {
	switch command {
	case "1":
		return a.addStudent(ctx)
	case "2":
		return a.selectStudent(ctx)
	case "3":
		return a.createRubric(ctx)
	case "4":
		return a.selectRubric(ctx)
	case "5":
		return a.showRubric(ctx)
	case "6", "7", "8", "9", "10", "11", "12", "13", "14":
		if a.studentID == uuid.Nil || a.rubricID == uuid.Nil {
			return fmt.Errorf("сначала выберите студента и рубрикатор")
		}
	default:
		return fmt.Errorf("неизвестная команда %q", command)
	}
	switch command {
	case "6":
		if _, err := a.forms.Create(ctx, a.studentID, a.rubricID); err != nil {
			return err
		}
		fmt.Fprintln(a.output, "Форма создана: Draft.")
	case "7":
		return a.showForm(ctx)
	case "8", "9":
		return a.enterEvaluation(ctx, command == "9")
	case "10":
		criterion, err := a.selectCriterion(ctx)
		if err != nil {
			return err
		}
		if err := a.forms.RemoveEvaluation(ctx, a.studentID, a.rubricID, criterion); err != nil {
			return err
		}
		fmt.Fprintln(a.output, "Оценка отменена.")
	case "11":
		if err := a.forms.Complete(ctx, a.studentID, a.rubricID); err != nil {
			return err
		}
		fmt.Fprintln(a.output, "Форма завершена: Completed.")
	case "12":
		if err := a.forms.Reopen(ctx, a.studentID, a.rubricID); err != nil {
			return err
		}
		fmt.Fprintln(a.output, "Форма открыта: Draft.")
	case "13":
		score, err := a.results.Calculate(ctx, a.studentID, a.rubricID)
		if err != nil {
			return err
		}
		fmt.Fprintf(a.output, "Итог: %g / 100\n", score.Value())
	case "14":
		return a.export(ctx)
	}
	return nil
}

func (a *App) addStudent(ctx context.Context) error {
	name, err := a.read("Имя студента: ")
	if err != nil {
		return err
	}
	student, err := a.students.Add(ctx, name)
	if err != nil {
		return err
	}
	a.studentID, a.studentName = student.ID(), student.Name()
	fmt.Fprintln(a.output, "Студент добавлен и выбран.")
	return nil
}

func (a *App) selectStudent(ctx context.Context) error {
	students, err := a.students.List(ctx)
	if err != nil {
		return err
	}
	if len(students) == 0 {
		return fmt.Errorf("студентов пока нет")
	}
	for i, student := range students {
		fmt.Fprintf(a.output, "%d. %s [%s]\n", i+1, student.Name(), student.ID())
	}
	index, err := a.choose(len(students))
	if err != nil {
		return err
	}
	a.studentID, a.studentName = students[index].ID(), students[index].Name()
	return nil
}

func (a *App) createRubric(ctx context.Context) error {
	name, err := a.read("Название рубрикатора: ")
	if err != nil {
		return err
	}
	count, err := a.integer("Количество критериев: ")
	if err != nil {
		return err
	}
	if count <= 0 {
		return fmt.Errorf("количество критериев должно быть положительным")
	}
	var criteria []application.CriterionInput
	for i := 0; i < count; i++ {
		fmt.Fprintf(a.output, "Критерий %d:\n", i+1)
		criterionName, err := a.read("Название: ")
		if err != nil {
			return err
		}
		maximum, err := a.number("Максимальный балл: ")
		if err != nil {
			return err
		}
		weight, err := a.integer("Вес в целых процентах (сумма всех весов — 100): ")
		if err != nil {
			return err
		}
		criteria = append(criteria, application.CriterionInput{Name: criterionName, MaxScore: maximum, Weight: weight})
	}
	rubric, err := a.rubrics.Create(ctx, name, criteria)
	if err != nil {
		return err
	}
	a.rubricID, a.rubricName = rubric.ID(), rubric.Name()
	fmt.Fprintln(a.output, "Рубрикатор создан и выбран.")
	return nil
}

func (a *App) selectRubric(ctx context.Context) error {
	rubrics, err := a.rubrics.List(ctx)
	if err != nil {
		return err
	}
	if len(rubrics) == 0 {
		return fmt.Errorf("рубрикаторов пока нет")
	}
	for i, rubric := range rubrics {
		fmt.Fprintf(a.output, "%d. %s [%s]\n", i+1, rubric.Name(), rubric.ID())
	}
	index, err := a.choose(len(rubrics))
	if err != nil {
		return err
	}
	a.rubricID, a.rubricName = rubrics[index].ID(), rubrics[index].Name()
	return nil
}

func (a *App) showRubric(ctx context.Context) error {
	if a.rubricID == uuid.Nil {
		return fmt.Errorf("сначала выберите рубрикатор")
	}
	rubric, err := a.rubrics.Get(ctx, a.rubricID)
	if err != nil {
		return err
	}
	for i, criterion := range rubric.Criteria() {
		fmt.Fprintf(a.output, "%d. %s: максимум %g, вес %d%%\n", i+1, criterion.Name(), criterion.MaxScore().Value(), criterion.Weight())
	}
	return nil
}

func (a *App) showForm(ctx context.Context) error {
	form, err := a.forms.Get(ctx, a.studentID, a.rubricID)
	if err != nil {
		return err
	}
	fmt.Fprintf(a.output, "Статус: %s\n", form.Status())
	evaluations := form.Evaluations()
	for _, criterion := range form.Rubric().Criteria() {
		found := false
		for _, evaluation := range evaluations {
			if evaluation.Criterion().Name() == criterion.Name() {
				fmt.Fprintf(a.output, "%s: %g / %g; вес %d%%; вклад %g; комментарий: %s\n",
					criterion.Name(), evaluation.Score().Value(), criterion.MaxScore().Value(), criterion.Weight(), evaluation.Contribution().Value(), evaluation.Comment)
				found = true
				break
			}
		}
		if !found {
			fmt.Fprintf(a.output, "%s: не оценено (максимум %g, вес %d%%)\n", criterion.Name(), criterion.MaxScore().Value(), criterion.Weight())
		}
	}
	return nil
}

func (a *App) selectCriterion(ctx context.Context) (string, error) {
	form, err := a.forms.Get(ctx, a.studentID, a.rubricID)
	if err != nil {
		return "", err
	}
	criteria := form.Rubric().Criteria()
	for i, criterion := range criteria {
		fmt.Fprintf(a.output, "%d. %s (максимум %g)\n", i+1, criterion.Name(), criterion.MaxScore().Value())
	}
	index, err := a.choose(len(criteria))
	if err != nil {
		return "", err
	}
	return criteria[index].Name(), nil
}

func (a *App) enterEvaluation(ctx context.Context, update bool) error {
	criterion, err := a.selectCriterion(ctx)
	if err != nil {
		return err
	}
	score, err := a.number("Балл: ")
	if err != nil {
		return err
	}
	comment, err := a.read("Комментарий (можно оставить пустым): ")
	if err != nil {
		return err
	}
	if update {
		err = a.forms.UpdateEvaluation(ctx, a.studentID, a.rubricID, criterion, score, comment)
	} else {
		err = a.forms.AddEvaluation(ctx, a.studentID, a.rubricID, criterion, score, comment)
	}
	if err != nil {
		return err
	}
	fmt.Fprintln(a.output, "Оценка сохранена.")
	return nil
}

func (a *App) export(ctx context.Context) error {
	if _, err := a.results.Calculate(ctx, a.studentID, a.rubricID); err != nil {
		return err
	}
	path, err := a.read("Путь к новому CSV-файлу: ")
	if err != nil {
		return err
	}
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0644)
	if err != nil {
		return fmt.Errorf("создать CSV-файл: %w", err)
	}
	exportErr := a.results.Export(ctx, a.studentID, a.rubricID, file)
	closeErr := file.Close()
	if err := errors.Join(exportErr, closeErr); err != nil {
		return fmt.Errorf("сохранить CSV: %w", err)
	}
	fmt.Fprintf(a.output, "CSV сохранён: %s\n", path)
	return nil
}

func (a *App) read(prompt string) (string, error) {
	if _, err := fmt.Fprint(a.output, prompt); err != nil {
		return "", err
	}
	if !a.input.Scan() {
		if err := a.input.Err(); err != nil {
			return "", fmt.Errorf("прочитать ввод: %w", err)
		}
		return "", io.EOF
	}
	return strings.TrimSpace(a.input.Text()), nil
}

func (a *App) integer(prompt string) (int, error) {
	text, err := a.read(prompt)
	if err != nil {
		return 0, err
	}
	value, err := strconv.Atoi(text)
	if err != nil {
		return 0, fmt.Errorf("ожидалось целое число, получено %q", text)
	}
	return value, nil
}

func (a *App) number(prompt string) (float64, error) {
	text, err := a.read(prompt)
	if err != nil {
		return 0, err
	}
	value, err := strconv.ParseFloat(strings.ReplaceAll(text, ",", "."), 64)
	if err != nil {
		return 0, fmt.Errorf("ожидалось число, получено %q", text)
	}
	return value, nil
}

func (a *App) choose(count int) (int, error) {
	number, err := a.integer("Номер: ")
	if err != nil {
		return 0, err
	}
	if number < 1 || number > count {
		return 0, fmt.Errorf("выберите номер от 1 до %d", count)
	}
	return number - 1, nil
}
