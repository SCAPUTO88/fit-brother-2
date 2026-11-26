package catalog

import (
	"strings"

	domainCatalog "github.com/SCAPUTO88/fit-brother/internal/domain/catalog"
)

var musclePTMap = map[string]string{
	"abdominals":  "Abdômen",
	"abductors":   "Abdutores",
	"adductors":   "Adutores",
	"biceps":      "Bíceps",
	"calves":      "Panturrilhas",
	"chest":       "Peitoral",
	"forearms":    "Antebraços",
	"glutes":      "Glúteos",
	"hamstrings":  "Posteriores da Coxa",
	"lats":        "Dorsais",
	"lower back":  "Lombar",
	"middle back": "Costas Média",
	"neck":        "Pescoço",
	"quadriceps":  "Quadríceps",
	"shoulders":   "Ombros",
	"traps":       "Trapézio",
	"triceps":     "Tríceps",
	"cardio":      "Cardio",
	"full body":   "Corpo Inteiro",
}

var equipmentPTMap = map[string]string{
	"barbell":       "Barra",
	"dumbbell":      "Halteres",
	"cable":         "Polia / Cabo",
	"machine":       "Máquina",
	"kettlebells":   "Kettlebell",
	"bands":         "Elásticos",
	"medicine ball": "Medicine Ball",
	"exercise ball": "Bola Suíça",
	"foam roll":     "Rolo de Liberação",
	"e-z curl bar":  "Barra W",
	"body only":     "Peso Corporal",
	"other":         "Outros",
}

var categoryPTMap = map[string]string{
	"strength":              "Força / Hipertrofia",
	"stretching":            "Alongamento",
	"cardio":                "Cardio",
	"plyometrics":           "Pliometria",
	"powerlifting":          "Powerlifting",
	"olympic weightlifting": "Levantamento Olímpico",
	"strongman":             "Strongman",
}

var knownExerciseNamesPT = map[string]string{
	"barbell bench press - medium grip":         "Supino Reto com Barra",
	"barbell incline bench press - medium grip": "Supino Inclinado com Barra",
	"dumbbell flyes":                            "Crucifixo Reto com Halteres",
	"pushups":                                   "Flexão de Braços",
	"cable crossover":                           "Crucifixo na Polia Alta (Crossover)",
	"wide-grip lat pulldown":                    "Puxada Frontal Aberta na Polia",
	"bent over barbell row":                     "Remada Curvada com Barra",
	"one-arm dumbbell row":                      "Remada Unilateral com Haltere (Serrote)",
	"pullups":                                   "Barra Fixa Pronada",
	"seated cable rows":                         "Remada Baixa na Polia",
	"barbell full squat":                        "Agachamento Livre com Barra",
	"leg press":                                 "Leg Press 45°",
	"leg extensions":                            "Cadeira Extensora",
	"dumbbell lunges":                           "Avanço / Passada com Halteres",
	"romanian deadlift with dumbbells":          "Stiff / Levantamento Romeno com Halteres",
	"lying leg curls":                           "Mesa Flexora",
	"barbell deadlift":                          "Levantamento Terra com Barra",
	"barbell hip thrust":                        "Elevação Pélvica com Barra",
	"standing calf raises":                      "Panturrilha em Pé",
	"dumbbell shoulder press":                   "Desenvolvimento com Halteres",
	"side lateral raise":                        "Elevação Lateral com Halteres",
	"face pull":                                 "Face Pull na Polia",
	"dumbbell bicep curl":                       "Rosca Direta com Halteres",
	"ez-bar curl":                               "Rosca Direta com Barra W",
	"triceps pushdown":                          "Tríceps Pulley (Barra/Corda)",
	"tricep dumbbell kickback":                  "Tríceps Coice com Haltere",
	"plank":                                     "Prancha Isométrica Abdominal",
	"crunches":                                  "Abdominal Supra no Solo",
	"kettlebell swing":                          "Swing com Kettlebell",
	"band pull-apart":                           "Abertura com Elástico (Band Pull-Apart)",
	"barbell shrug":                             "Encolhimento de Ombros com Barra",
	"jump rope":                                 "Pular Corda (Cardio)",
	"barbell squat":                             "Agachamento com Barra",
	"barbell curl":                              "Rosca Direta com Barra",
	"hammer curls":                              "Rosca Martelo com Halteres",
	"incline dumbbell press":                    "Supino Inclinado com Halteres",
	"dumbbell bench press":                      "Supino Reto com Halteres",
}

var phraseReplacementsPT = []struct {
	en string
	pt string
}{
	{"starting position", "posição inicial"},
	{"shoulder-width", "largura dos ombros"},
	{"shoulder width", "largura dos ombros"},
	{"shoulder blades", "escápulas"},
	{"upper chest", "parte superior do peitoral"},
	{"middle chest", "meio do peitoral"},
	{"lower back", "lombar"},
	{"flat bench", "banco reto"},
	{"incline bench", "banco inclinado"},
	{"kettlebell", "kettlebell"},
	{"dumbbells", "halteres"},
	{"dumbbell", "haltere"},
	{"barbell", "barra"},
	{"elbows", "cotovelos"},
	{"elbow", "cotovelo"},
	{"knees", "joelhos"},
	{"knee", "joelho"},
	{"hips", "quadril"},
	{"hip", "quadril"},
	{"feet", "pés"},
	{"foot", "pé"},
	{"hands", "mãos"},
	{"hand", "mão"},
	{"chest", "peito"},
	{"shoulders", "ombros"},
	{"back", "costas"},
	{"arms", "braços"},
	{"legs", "pernas"},
	{"floor", "solo"},
	{"slowly", "lentamente"},
	{"control", "controle"},
	{"breathe out", "expire"},
	{"breathe in", "inspire"},
	{"inhale", "inspire"},
	{"exhale", "expire"},
}

func TranslateMusclePT(code string) string {
	norm := domainCatalog.NormalizeCode(code)
	if v, ok := musclePTMap[norm]; ok {
		return v
	}
	if norm == "" {
		return "Geral"
	}
	return strings.TrimSpace(code)
}

func TranslateEquipmentPT(code string) string {
	norm := domainCatalog.NormalizeCode(code)
	if v, ok := equipmentPTMap[norm]; ok {
		return v
	}
	if norm == "" {
		return "Peso Corporal"
	}
	return strings.TrimSpace(code)
}

func TranslateCategoryPT(code string) string {
	norm := domainCatalog.NormalizeCode(code)
	if v, ok := categoryPTMap[norm]; ok {
		return v
	}
	if norm == "" {
		return "Força / Hipertrofia"
	}
	return strings.TrimSpace(code)
}

func TranslateExerciseNamePT(nameEN, muscleCode, equipmentCode string) string {
	norm := strings.ToLower(strings.TrimSpace(nameEN))
	if pt, ok := knownExerciseNamesPT[norm]; ok {
		return pt
	}
	trimmed := strings.TrimSpace(nameEN)
	if trimmed == "" {
		return "Exercício (" + TranslateMusclePT(muscleCode) + " - " + TranslateEquipmentPT(equipmentCode) + ")"
	}
	return trimmed
}

func TranslateInstructionsPT(instructionsEN []string) []string {
	if len(instructionsEN) == 0 {
		return []string{}
	}
	out := make([]string, 0, len(instructionsEN))
	for _, step := range instructionsEN {
		trimmed := strings.TrimSpace(step)
		if trimmed == "" {
			continue
		}
		translated := trimmed
		for _, rep := range phraseReplacementsPT {
			translated = replaceCaseInsensitive(translated, rep.en, rep.pt)
		}
		out = append(out, translated)
	}
	return out
}

func replaceCaseInsensitive(input, target, replacement string) string {
	lowerInput := strings.ToLower(input)
	lowerTarget := strings.ToLower(target)
	if !strings.Contains(lowerInput, lowerTarget) {
		return input
	}

	var b strings.Builder
	i := 0
	for {
		idx := strings.Index(lowerInput[i:], lowerTarget)
		if idx == -1 {
			b.WriteString(input[i:])
			break
		}
		b.WriteString(input[i : i+idx])
		b.WriteString(replacement)
		i += idx + len(target)
	}
	return b.String()
}
