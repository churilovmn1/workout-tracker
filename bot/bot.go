package bot

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/churilovmn1/workout-tracker/internal/models"
	"github.com/churilovmn1/workout-tracker/internal/repository"
	"github.com/churilovmn1/workout-tracker/internal/service"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// Bot wraps the Telegram bot with application services.
type Bot struct {
	api             *tgbotapi.BotAPI
	userRepo        *repository.UserRepository
	workoutService  *service.WorkoutService
	exerciseService *service.ExerciseService
	templateService *service.TemplateService
	sessions        map[int64]*session
	mu              sync.Mutex
}

type sessionState int

const (
	stateNone sessionState = iota
	stateAwaitTitle
	stateAwaitExercises
)

type session struct {
	state    sessionState
	workout  *models.Workout
	userID   int
}

// New creates a new Bot instance.
func New(
	token string,
	userRepo *repository.UserRepository,
	workoutService *service.WorkoutService,
	exerciseService *service.ExerciseService,
	templateService *service.TemplateService,
) (*Bot, error) {
	api, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		return nil, fmt.Errorf("create bot api: %w", err)
	}

	log.Printf("telegram bot authorized: %s", api.Self.UserName)

	return &Bot{
		api:             api,
		userRepo:        userRepo,
		workoutService:  workoutService,
		exerciseService: exerciseService,
		templateService: templateService,
		sessions:        make(map[int64]*session),
	}, nil
}

// Start begins polling for updates and handling commands.
func (b *Bot) Start(ctx context.Context) {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates := b.api.GetUpdatesChan(u)

	for {
		select {
		case <-ctx.Done():
			b.api.StopReceivingUpdates()
			return
		case update := <-updates:
			if update.Message == nil {
				continue
			}
			b.handleMessage(ctx, update.Message)
		}
	}
}

func (b *Bot) handleMessage(ctx context.Context, msg *tgbotapi.Message) {
	telegramID := msg.From.ID

	if msg.IsCommand() {
		b.handleCommand(ctx, msg, telegramID)
		return
	}

	b.mu.Lock()
	sess, ok := b.sessions[telegramID]
	b.mu.Unlock()

	if !ok {
		b.send(msg.Chat.ID, "Используй /help чтобы увидеть доступные команды.")
		return
	}

	b.handleSession(ctx, msg, sess, telegramID)
}

func (b *Bot) handleCommand(ctx context.Context, msg *tgbotapi.Message, telegramID int64) {
	switch msg.Command() {
	case "start":
		b.cmdStart(ctx, msg, telegramID)
	case "help":
		b.cmdHelp(msg)
	case "newworkout":
		b.cmdNewWorkout(ctx, msg, telegramID)
	case "workouts":
		b.cmdWorkouts(ctx, msg, telegramID)
	case "pr":
		b.cmdPR(ctx, msg, telegramID)
	case "stats":
		b.cmdStats(ctx, msg, telegramID)
	case "exercises":
		b.cmdExercises(ctx, msg)
	case "cancel":
		b.mu.Lock()
		delete(b.sessions, telegramID)
		b.mu.Unlock()
		b.send(msg.Chat.ID, "Действие отменено.")
	default:
		b.send(msg.Chat.ID, "Неизвестная команда. /help для справки.")
	}
}

func (b *Bot) cmdStart(ctx context.Context, msg *tgbotapi.Message, telegramID int64) {
	_, err := b.userRepo.GetByTelegramID(ctx, telegramID)
	if err != nil {
		tgID := telegramID
		user := &models.User{
			Login:        fmt.Sprintf("tg_%d", telegramID),
			Email:        fmt.Sprintf("tg_%d@telegram.local", telegramID),
			PasswordHash: "-",
			Role:         models.RoleUser,
			TelegramID:   &tgID,
		}
		_, err = b.userRepo.Create(ctx, user)
		if err != nil {
			b.send(msg.Chat.ID, "Ошибка регистрации. Попробуй позже.")
			return
		}
	}

	b.send(msg.Chat.ID, fmt.Sprintf(
		"Привет, %s! Я бот для трекинга тренировок.\n\n"+
			"Используй /help чтобы увидеть команды.",
		msg.From.FirstName,
	))
}

func (b *Bot) cmdHelp(msg *tgbotapi.Message) {
	text := `Доступные команды:

/newworkout — записать новую тренировку
/workouts — список последних тренировок
/pr — личные рекорды
/stats — статистика за неделю
/exercises — каталог упражнений
/cancel — отменить текущее действие
/help — эта справка`

	b.send(msg.Chat.ID, text)
}

func (b *Bot) cmdNewWorkout(ctx context.Context, msg *tgbotapi.Message, telegramID int64) {
	user, err := b.userRepo.GetByTelegramID(ctx, telegramID)
	if err != nil {
		b.send(msg.Chat.ID, "Сначала нажми /start для регистрации.")
		return
	}

	b.mu.Lock()
	b.sessions[telegramID] = &session{
		state:  stateAwaitTitle,
		userID: user.ID,
		workout: &models.Workout{
			UserID: user.ID,
			Date:   time.Now(),
		},
	}
	b.mu.Unlock()

	b.send(msg.Chat.ID, "Введи название тренировки (например: Грудь + Трицепс):")
}

func (b *Bot) handleSession(ctx context.Context, msg *tgbotapi.Message, sess *session, telegramID int64) {
	switch sess.state {
	case stateAwaitTitle:
		sess.workout.Title = msg.Text
		sess.state = stateAwaitExercises

		b.send(msg.Chat.ID,
			"Теперь добавляй упражнения в формате:\n"+
				"`ID подходы повторы вес`\n\n"+
				"Например: `1 4 10 80`\n"+
				"(упражнение #1, 4 подхода, 10 повторов, 80 кг)\n\n"+
				"Отправь /done когда закончишь, /exercises чтобы посмотреть ID упражнений.")

	case stateAwaitExercises:
		if msg.Text == "/done" {
			id, err := b.workoutService.Create(ctx, sess.workout)
			if err != nil {
				b.send(msg.Chat.ID, "Ошибка сохранения: "+err.Error())
				return
			}

			b.mu.Lock()
			delete(b.sessions, telegramID)
			b.mu.Unlock()

			b.send(msg.Chat.ID, fmt.Sprintf(
				"Тренировка \"%s\" сохранена! (ID: %d)\nУпражнений: %d",
				sess.workout.Title, id, len(sess.workout.Exercises),
			))
			return
		}

		parts := strings.Fields(msg.Text)
		if len(parts) != 4 {
			b.send(msg.Chat.ID, "Формат: `ID подходы повторы вес`\nПример: `1 4 10 80`")
			return
		}

		exID, err1 := strconv.Atoi(parts[0])
		sets, err2 := strconv.Atoi(parts[1])
		reps, err3 := strconv.Atoi(parts[2])
		weight, err4 := strconv.ParseFloat(parts[3], 64)

		if err1 != nil || err2 != nil || err3 != nil || err4 != nil {
			b.send(msg.Chat.ID, "Все значения должны быть числами. Формат: `ID подходы повторы вес`")
			return
		}

		sess.workout.Exercises = append(sess.workout.Exercises, models.WorkoutExercise{
			ExerciseID: exID,
			Sets:       sets,
			Reps:       reps,
			WeightKg:   weight,
		})

		b.send(msg.Chat.ID, fmt.Sprintf(
			"Добавлено! (упражнение #%d: %dx%d, %.1f кг)\nЕщё упражнение или /done для сохранения.",
			exID, sets, reps, weight,
		))
	}
}

func (b *Bot) cmdWorkouts(ctx context.Context, msg *tgbotapi.Message, telegramID int64) {
	user, err := b.userRepo.GetByTelegramID(ctx, telegramID)
	if err != nil {
		b.send(msg.Chat.ID, "Сначала нажми /start для регистрации.")
		return
	}

	workouts, err := b.workoutService.ListByUser(ctx, user.ID)
	if err != nil {
		b.send(msg.Chat.ID, "Ошибка загрузки тренировок.")
		return
	}

	if len(workouts) == 0 {
		b.send(msg.Chat.ID, "У тебя пока нет тренировок. Создай первую: /newworkout")
		return
	}

	var sb strings.Builder
	sb.WriteString("Последние тренировки:\n\n")
	limit := len(workouts)
	if limit > 10 {
		limit = 10
	}
	for _, w := range workouts[:limit] {
		sb.WriteString(fmt.Sprintf("• %s — %s (%d мин)\n",
			w.Date.Format("02.01.2006"), w.Title, w.DurationMinutes))
	}

	b.send(msg.Chat.ID, sb.String())
}

func (b *Bot) cmdPR(ctx context.Context, msg *tgbotapi.Message, telegramID int64) {
	user, err := b.userRepo.GetByTelegramID(ctx, telegramID)
	if err != nil {
		b.send(msg.Chat.ID, "Сначала нажми /start для регистрации.")
		return
	}

	records, err := b.workoutService.GetPersonalRecords(ctx, user.ID)
	if err != nil {
		b.send(msg.Chat.ID, "Ошибка загрузки рекордов.")
		return
	}

	if len(records) == 0 {
		b.send(msg.Chat.ID, "Пока нет рекордов. Запиши тренировку: /newworkout")
		return
	}

	var sb strings.Builder
	sb.WriteString("Личные рекорды:\n\n")
	for _, r := range records {
		ex, err := b.exerciseService.GetByID(ctx, r.ExerciseID)
		name := fmt.Sprintf("Упражнение #%d", r.ExerciseID)
		if err == nil {
			name = ex.Name
		}
		sb.WriteString(fmt.Sprintf("• %s — %.1f кг (%dx%d)\n",
			name, r.WeightKg, r.Sets, r.Reps))
	}

	b.send(msg.Chat.ID, sb.String())
}

func (b *Bot) cmdStats(ctx context.Context, msg *tgbotapi.Message, telegramID int64) {
	user, err := b.userRepo.GetByTelegramID(ctx, telegramID)
	if err != nil {
		b.send(msg.Chat.ID, "Сначала нажми /start для регистрации.")
		return
	}

	volume, err := b.workoutService.GetWeeklyVolume(ctx, user.ID)
	if err != nil {
		b.send(msg.Chat.ID, "Ошибка загрузки статистики.")
		return
	}

	workouts, err := b.workoutService.ListByUser(ctx, user.ID)
	if err != nil {
		b.send(msg.Chat.ID, "Ошибка загрузки тренировок.")
		return
	}

	weekCount := 0
	weekAgo := time.Now().AddDate(0, 0, -7)
	for _, w := range workouts {
		if w.Date.After(weekAgo) {
			weekCount++
		}
	}

	streak := calcStreak(workouts)

	b.send(msg.Chat.ID, fmt.Sprintf(
		"Статистика:\n\n"+
			"Тренировок за неделю: %d\n"+
			"Объём за неделю: %.0f кг\n"+
			"Серия тренировок: %d дн.\n"+
			"Всего тренировок: %d",
		weekCount, volume, streak, len(workouts),
	))
}

func (b *Bot) cmdExercises(ctx context.Context, msg *tgbotapi.Message) {
	exercises, err := b.exerciseService.List(ctx, "")
	if err != nil {
		b.send(msg.Chat.ID, "Ошибка загрузки упражнений.")
		return
	}

	if len(exercises) == 0 {
		b.send(msg.Chat.ID, "Каталог упражнений пуст. Админ должен добавить упражнения.")
		return
	}

	var sb strings.Builder
	sb.WriteString("Каталог упражнений:\n\n")
	for _, e := range exercises {
		sb.WriteString(fmt.Sprintf("#%d — %s (%s)\n", e.ID, e.Name, e.MuscleGroup))
	}

	b.send(msg.Chat.ID, sb.String())
}

func (b *Bot) send(chatID int64, text string) {
	reply := tgbotapi.NewMessage(chatID, text)
	reply.ParseMode = "Markdown"
	if _, err := b.api.Send(reply); err != nil {
		log.Printf("bot send error: %v", err)
	}
}

func calcStreak(workouts []models.Workout) int {
	if len(workouts) == 0 {
		return 0
	}

	streak := 0
	check := time.Now().Truncate(24 * time.Hour)

	for _, w := range workouts {
		wDate := w.Date.Truncate(24 * time.Hour)
		diff := check.Sub(wDate)

		if diff > 48*time.Hour {
			break
		}
		if diff >= 0 {
			streak++
			check = wDate.Add(-24 * time.Hour)
		}
	}

	return streak
}
