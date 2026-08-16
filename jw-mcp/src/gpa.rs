use crate::parse::Grade;
use serde::{Deserialize, Serialize};
use std::collections::HashMap;

#[derive(Debug, Clone, Serialize, Deserialize, PartialEq)]
pub struct Gpa {
    pub average_gpa: f64,
    pub average_score: f64,
    pub basic_score: f64,
}

#[derive(Debug, Clone, Serialize, Deserialize, PartialEq)]
pub struct CreditSummary {
    pub required_credits: f64,
    pub earned_required_credits: f64,
    pub total_credits: f64,
    pub required_course_count: usize,
    pub gpa: Gpa,
}

pub fn distinct_grades(grades: &[Grade]) -> Vec<Grade> {
    let mut map: HashMap<String, Grade> = HashMap::new();
    for g in grades {
        let key = format!("{}|{}|{}", g.serial_no, g.code, g.term);
        match map.get(&key) {
            Some(existing) if existing.flag == "缓考" && g.flag != "缓考" => {
                map.insert(key, g.clone());
            }
            None => {
                map.insert(key, g.clone());
            }
            _ => {}
        }
    }
    map.into_values().collect()
}

pub fn calculate_gpa(grades: &[Grade]) -> Gpa {
    let distinct = distinct_grades(grades);
    let mut sum_score = 0.0;
    let mut sum_gp = 0.0;
    let mut sum_credit = 0.0;
    let mut num2 = 0.0;
    let mut sum_score2 = 0.0;
    let mut sum_credit2 = 0.0;

    for g in &distinct {
        if g.property != "必修" {
            continue;
        }
        if g.flag == "缓考" {
            continue;
        }
        let score_text = g.score.as_str();
        if g.status == 0 {
            sum_score2 += map_grade_to_score_for_basic(score_text) * g.credit;
            sum_credit2 += g.credit;
        }

        let numeric = parse_numeric(score_text);
        if let Some(numeric_score) = numeric {
            if g.status == 0 && numeric_score >= 59.9 {
                sum_score += numeric_score;
                sum_gp += course_gp(g, score_text) * g.credit;
                sum_credit += g.credit;
                num2 += 1.0;
            } else if g.status == 1 && numeric_score >= 59.9 {
                sum_score += 60.0;
                sum_gp += course_gp(g, score_text) * 1.0;
                sum_credit += g.credit;
                num2 += 1.0;
            } else {
                sum_credit += g.credit;
                num2 += 1.0;
                if g.status != 1 || numeric_score > 59.9 {
                    sum_score += numeric_score;
                }
            }
        } else if g.status == 0 {
            let gp = course_gp(g, score_text);
            sum_score += gp * 10.0 + 50.0;
            sum_gp += gp * g.credit;
            sum_credit += g.credit;
            num2 += 1.0;
        } else if g.status == 1 && (score_text == "及格" || score_text == "合格") {
            let gp = course_gp(g, score_text);
            sum_score += 60.0;
            sum_gp += gp * 1.0;
            sum_credit += g.credit;
            num2 += 1.0;
        } else {
            sum_credit += g.credit;
            num2 += 1.0;
        }
    }

    let gpa = if sum_credit != 0.0 { sum_gp / sum_credit } else { 0.0 };
    let apf = if num2 != 0.0 { sum_score / num2 } else { 0.0 };
    let basic = if sum_credit2 != 0.0 {
        sum_score2 / sum_credit2
    } else {
        0.0
    };

    Gpa {
        average_gpa: round3(nan_to_zero(gpa)),
        average_score: round3(nan_to_zero(apf)),
        basic_score: round3(nan_to_zero(basic)),
    }
}

pub fn credit_summary(grades: &[Grade]) -> CreditSummary {
    let distinct = distinct_grades(grades);
    let mut required_credits = 0.0;
    let mut earned_required = 0.0;
    let mut total_credits = 0.0;
    let mut required_count = 0;
    for g in &distinct {
        if g.flag == "缓考" {
            continue;
        }
        total_credits += g.credit;
        if g.property == "必修" {
            required_credits += g.credit;
            required_count += 1;
            if course_passed(g) {
                earned_required += g.credit;
            }
        }
    }
    CreditSummary {
        required_credits: round3(required_credits),
        earned_required_credits: round3(earned_required),
        total_credits: round3(total_credits),
        required_course_count: required_count,
        gpa: calculate_gpa(grades),
    }
}

fn course_passed(g: &Grade) -> bool {
    if let Some(n) = parse_numeric(&g.score) {
        return n >= 59.9;
    }
    matches!(g.score.as_str(), "及格" | "合格" | "中" | "良" | "优")
}

fn course_gp(g: &Grade, score_text: &str) -> f64 {
    if !g.gpa.is_nan() && g.gpa > 0.0 {
        g.gpa
    } else {
        handel_gp(score_text)
    }
}

fn handel_gp(score_text: &str) -> f64 {
    match score_text {
        "不及格" | "不合格" => 0.0,
        "及格" | "合格" => 1.0,
        "中" => 2.0,
        "良" => 3.0,
        "优" => 4.0,
        _ => {
            if let Some(score) = parse_numeric(score_text) {
                let raw = round3((score - 50.0) / 10.0);
                if raw <= 0.1 {
                    0.0
                } else {
                    raw
                }
            } else {
                0.0
            }
        }
    }
}

fn map_grade_to_score_for_basic(score_text: &str) -> f64 {
    match score_text {
        "不及格" | "不合格" => 50.0,
        "及格" | "合格" => 60.0,
        "中" => 70.0,
        "良" => 80.0,
        "优" => 90.0,
        _ => parse_numeric(score_text).unwrap_or(0.0),
    }
}

fn parse_numeric(s: &str) -> Option<f64> {
    s.trim().parse().ok()
}

fn round3(v: f64) -> f64 {
    (v * 1000.0).round() / 1000.0
}

fn nan_to_zero(v: f64) -> f64 {
    if v.is_nan() {
        0.0
    } else {
        v
    }
}
