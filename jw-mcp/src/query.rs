use crate::client::{build_http_client, endpoints_for, fetch_text};
use crate::gpa::{credit_summary, CreditSummary, Gpa};
use crate::parse::{
    filter_grades_by_term, parse_courses_from_html, parse_grades_from_html, parse_level_grades_from_html,
    parse_student_info_from_html, Grade, LevelGrade, StudentInfo, WeekSchedule,
};
use crate::session::Session;
use anyhow::{Context, Result};
use wreq::Method;
use serde::Serialize;
use std::collections::BTreeMap;

#[derive(Debug, Serialize)]
pub struct GradesPayload {
    pub term: String,
    pub count: usize,
    pub grades: Vec<Grade>,
    pub gpa: Gpa,
    pub credits: CreditSummary,
    pub by_term: BTreeMap<String, TermBucket>,
}

#[derive(Debug, Serialize)]
pub struct TermBucket {
    pub count: usize,
    pub gpa: Gpa,
    pub credits: CreditSummary,
}

pub async fn query_grades(session: &Session, term: Option<&str>) -> Result<GradesPayload> {
    let client = build_http_client()?;
    let ep = endpoints_for(session);
    let mut form: Vec<(&str, &str)> = vec![("kksj", ""), ("kcxz", ""), ("kcmc", ""), ("xsfs", "all")];
    if let Some(t) = term {
        if !t.is_empty() {
            form[0] = ("kksj", t);
        }
    }
    let html = fetch_text(&client, session, Method::POST, &ep.grade_url, Some(&form))
        .await
        .context("拉取成绩页失败")?;
    let all = parse_grades_from_html(&html)?;
    let selected = match term {
        Some(t) if !t.is_empty() => filter_grades_by_term(&all, t),
        _ => all.clone(),
    };
    Ok(build_grades_payload(term.unwrap_or(""), selected, &all))
}

pub fn build_grades_payload(term: &str, selected: Vec<Grade>, all: &[Grade]) -> GradesPayload {
    let mut by_term: BTreeMap<String, TermBucket> = BTreeMap::new();
    let mut grouped: BTreeMap<String, Vec<Grade>> = BTreeMap::new();
    for g in all {
        grouped.entry(g.term.clone()).or_default().push(g.clone());
    }
    for (t, gs) in grouped {
        by_term.insert(
            t,
            TermBucket {
                count: gs.len(),
                gpa: crate::gpa::calculate_gpa(&gs),
                credits: credit_summary(&gs),
            },
        );
    }
    GradesPayload {
        term: term.to_string(),
        count: selected.len(),
        gpa: crate::gpa::calculate_gpa(&selected),
        credits: credit_summary(&selected),
        grades: selected,
        by_term,
    }
}

pub async fn query_level_exams(session: &Session) -> Result<Vec<LevelGrade>> {
    let client = build_http_client()?;
    let ep = endpoints_for(session);
    let html = fetch_text(&client, session, Method::GET, &ep.level_url, None)
        .await
        .context("拉取等级考试页失败")?;
    Ok(parse_level_grades_from_html(&html)?)
}

pub async fn query_courses(session: &Session, term: &str, week: i32) -> Result<WeekSchedule> {
    if !(1..=20).contains(&week) {
        anyhow::bail!("周次必须在 1-20 之间");
    }
    if !regex::Regex::new(r"^\d{4}-\d{4}-[12]$").unwrap().is_match(term) {
        anyhow::bail!("学期格式错误，应为 2024-2025-1");
    }
    let client = build_http_client()?;
    let ep = endpoints_for(session);
    let week_s = week.to_string();
    let form = [("zc", week_s.as_str()), ("xnxq01id", term)];
    let html = fetch_text(&client, session, Method::POST, &ep.course_url, Some(&form))
        .await
        .context("拉取课表页失败")?;
    Ok(parse_courses_from_html(&html, week)?)
}

pub async fn query_student_info(session: &Session) -> Result<StudentInfo> {
    let client = build_http_client()?;
    let ep = endpoints_for(session);
    let form = [("kksj", ""), ("kcxz", ""), ("kcmc", ""), ("xsfs", "all")];
    let html = fetch_text(&client, session, Method::POST, &ep.grade_url, Some(&form)).await?;
    Ok(parse_student_info_from_html(&html, "")?)
}

pub async fn query_credits(session: &Session) -> Result<CreditSummary> {
    let payload = query_grades(session, None).await?;
    Ok(payload.credits)
}
