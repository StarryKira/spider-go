package service

import (
	"context"
	"spider-go/internal/cache"
	"spider-go/internal/common"
)

// GradeAnalysisService 成绩分析服务接口
type GradeAnalysisService interface {
	// GetRecentTermsGrades 获取最近三个学期的成绩分析
	GetRecentTermsGrades(ctx context.Context, uid int) (*TermsGradesAnalysis, error)
}

// gradeAnalysisServiceImpl 成绩分析服务实现
type gradeAnalysisServiceImpl struct {
	gradeService GradeService
	configCache  cache.ConfigCache
}

// NewGradeAnalysisService 创建成绩分析服务
func NewGradeAnalysisService(
	gradeService GradeService,
	configCache cache.ConfigCache,
) GradeAnalysisService {
	return &gradeAnalysisServiceImpl{
		gradeService: gradeService,
		configCache:  configCache,
	}
}

// TermGradesData 单个学期的成绩数据
type TermGradesData struct {
	Term   string  `json:"term"`   // 学期
	Grades []Grade `json:"grades"` // 成绩列表
	GPA    *GPA    `json:"gpa"`    // GPA统计
}

// TermsGradesAnalysis 多学期成绩分析
type TermsGradesAnalysis struct {
	CurrentTerm   string           `json:"current_term"`   // 当前学期
	TermsData     []TermGradesData `json:"terms_data"`     // 各学期数据
	OverallGPA    *GPA             `json:"overall_gpa"`    // 总体GPA
	TrendAnalysis *TrendAnalysis   `json:"trend_analysis"` // 趋势分析
}

// TrendAnalysis 趋势分析
type TrendAnalysis struct {
	GPATrend     string  `json:"gpa_trend"`      // GPA趋势：上升/下降/稳定
	ScoreTrend   string  `json:"score_trend"`    // 成绩趋势
	BestTerm     string  `json:"best_term"`      // 最好的学期
	BestTermGPA  float64 `json:"best_term_gpa"`  // 最好学期的GPA
	WorstTerm    string  `json:"worst_term"`     // 最差的学期
	WorstTermGPA float64 `json:"worst_term_gpa"` // 最差学期的GPA
}

// GetRecentTermsGrades 获取最近三个学期的成绩分析
func (s *gradeAnalysisServiceImpl) GetRecentTermsGrades(ctx context.Context, uid int) (*TermsGradesAnalysis, error) {
	// 1. 获取最近三个学期
	terms, err := s.configCache.GetPreviousTerms(ctx, 3)
	if err != nil {
		return nil, common.NewAppError(common.CodeInternalError, err.Error())
	}

	// 2. 获取所有成绩
	allGrades, overallGPA, err := s.gradeService.GetAllGrade(ctx, uid)
	if err != nil {
		return nil, err
	}

	// 3. 按学期分组成绩
	termsData := make([]TermGradesData, 0)
	for _, term := range terms {
		termGrades := s.filterGradesByTerm(allGrades, term)
		if len(termGrades) == 0 {
			// 如果该学期没有成绩，添加空数据
			termsData = append(termsData, TermGradesData{
				Term:   term,
				Grades: []Grade{},
				GPA:    &GPA{},
			})
			continue
		}

		// 计算该学期的 GPA
		termGPA, err := s.calculateTermGPA(termGrades)
		if err != nil {
			termGPA = &GPA{}
		}

		termsData = append(termsData, TermGradesData{
			Term:   term,
			Grades: termGrades,
			GPA:    termGPA,
		})
	}

	// 4. 趋势分析
	trendAnalysis := s.analyzeTrend(termsData)

	return &TermsGradesAnalysis{
		CurrentTerm:   terms[0],
		TermsData:     termsData,
		OverallGPA:    overallGPA,
		TrendAnalysis: trendAnalysis,
	}, nil
}

// filterGradesByTerm 按学期过滤成绩
func (s *gradeAnalysisServiceImpl) filterGradesByTerm(grades []Grade, term string) []Grade {
	filtered := make([]Grade, 0)
	for _, grade := range grades {
		if grade.Term == term {
			filtered = append(filtered, grade)
		}
	}
	return filtered
}

// calculateTermGPA 计算单个学期的 GPA（复用 GradeService 的逻辑）
func (s *gradeAnalysisServiceImpl) calculateTermGPA(grades []Grade) (*GPA, error) {
	// 创建一个临时的 gradeServiceImpl 来计算 GPA
	// 这里简化处理，直接调用已有的计算逻辑
	tempService := &gradeServiceImpl{}
	return tempService.calculateGPA(grades)
}

// analyzeTrend 分析成绩趋势
func (s *gradeAnalysisServiceImpl) analyzeTrend(termsData []TermGradesData) *TrendAnalysis {
	if len(termsData) < 2 {
		return &TrendAnalysis{
			GPATrend:   "数据不足",
			ScoreTrend: "数据不足",
		}
	}

	// 找出最好和最差的学期
	var bestTerm, worstTerm string
	var bestGPA, worstGPA float64 = -1, 999

	gpas := make([]float64, 0)
	for _, data := range termsData {
		if len(data.Grades) == 0 {
			continue
		}
		gpa := data.GPA.AverageGPA
		gpas = append(gpas, gpa)

		if gpa > bestGPA {
			bestGPA = gpa
			bestTerm = data.Term
		}
		if gpa < worstGPA {
			worstGPA = gpa
			worstTerm = data.Term
		}
	}

	// 分析趋势
	gpaTrend := "稳定"
	scoreTrend := "稳定"

	if len(gpas) >= 2 {
		// 比较最近两个学期
		if gpas[0] > gpas[1]+0.1 {
			gpaTrend = "上升"
			scoreTrend = "上升"
		} else if gpas[0] < gpas[1]-0.1 {
			gpaTrend = "下降"
			scoreTrend = "下降"
		}
	}

	return &TrendAnalysis{
		GPATrend:     gpaTrend,
		ScoreTrend:   scoreTrend,
		BestTerm:     bestTerm,
		BestTermGPA:  bestGPA,
		WorstTerm:    worstTerm,
		WorstTermGPA: worstGPA,
	}
}
