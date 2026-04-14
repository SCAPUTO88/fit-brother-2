package workout

type CreatePlanInput struct {
	TrainerID   string `json:"trainer_id" binding:"required,uuid"`
	StudentID   string `json:"student_id" binding:"required,uuid"`
	Title       string `json:"title" binding:"required,min=3"`
	Description string `json:"description"`
}

type CreateSessionInput struct {
	WorkoutPlanID string `json:"workout_plan_id" binding:"required,uuid"`
	Title         string `json:"title" binding:"required,min=2"`
	DayOfWeek     int    `json:"day_of_week" binding:"min=0,max=6"`
	Notes         string `json:"notes"`
}

type CreateExerciseInput struct {
	WorkoutSessionID string `json:"workout_session_id" binding:"required,uuid"`
	Name             string `json:"name" binding:"required,min=2"`
	Sets             int    `json:"sets" binding:"min=0"`
	Reps             string `json:"reps"`
	RestSeconds      int    `json:"rest_seconds"`
	Notes            string `json:"notes"`
	VideoURL         string `json:"video_url"`
	Position         int    `json:"position"`
}

type PlanOutput struct {
	ID          string `json:"id"`
	TrainerID   string `json:"trainer_id"`
	StudentID   string `json:"student_id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	IsActive    bool   `json:"is_active"`
	CreatedAt   string `json:"created_at"`
}

type SessionOutput struct {
	ID            string `json:"id"`
	WorkoutPlanID string `json:"workout_plan_id"`
	Title         string `json:"title"`
	DayOfWeek     int    `json:"day_of_week"`
	Notes         string `json:"notes"`
	CreatedAt     string `json:"created_at"`
}

type ExerciseOutput struct {
	ID               string `json:"id"`
	WorkoutSessionID string `json:"workout_session_id"`
	Name             string `json:"name"`
	Sets             int    `json:"sets"`
	Reps             string `json:"reps"`
	RestSeconds      int    `json:"rest_seconds"`
	Notes            string `json:"notes"`
	VideoURL         string `json:"video_url"`
	Position         int    `json:"position"`
}
