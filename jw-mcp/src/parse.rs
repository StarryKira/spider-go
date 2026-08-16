use regex::Regex;
use scraper::{Html, Selector};
use serde::{Deserialize, Serialize};

#[derive(Debug, Clone, Serialize, Deserialize, PartialEq)]
pub struct Grade {
    pub serial_no: String,
    pub term: String,
    pub code: String,
    pub subject: String,
    pub score: String,
    pub credit: f64,
    pub gpa: f64,
    pub status: i32,
    pub property: String,
    pub flag: String,
}

#[derive(Debug, Clone, Serialize, Deserialize, PartialEq)]
pub struct LevelGrade {
    pub no: String,
    pub course_name: String,
    pub level_grade: String,
    pub time: String,
}

#[derive(Debug, Clone, Serialize, Deserialize, PartialEq)]
pub struct Course {
    pub name: String,
    pub teacher: String,
    pub classroom: String,
    pub weekday: i32,
    pub start_period: i32,
    pub end_period: i32,
}

#[derive(Debug, Clone, Serialize, Deserialize, PartialEq)]
pub struct DaySchedule {
    pub weekday: i32,
    pub courses: Vec<Course>,
}

#[derive(Debug, Clone, Serialize, Deserialize, PartialEq)]
pub struct WeekSchedule {
    pub weekno: i32,
    pub days: Vec<DaySchedule>,
}

#[derive(Debug, Clone, Serialize, Deserialize, PartialEq, Default)]
pub struct StudentInfo {
    pub name: String,
    pub grade: String,
    pub class: String,
    pub major: String,
    pub college: String,
}

#[derive(Debug)]
pub enum ParseError {
    Unevaluated(String),
    MissingTable(String),
    PageError(String),
}

impl std::fmt::Display for ParseError {
    fn fmt(&self, f: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
        match self {
            Self::Unevaluated(s) | Self::MissingTable(s) | Self::PageError(s) => write!(f, "{s}"),
        }
    }
}

impl std::error::Error for ParseError {}

pub fn looks_like_login_page(html: &str) -> bool {
    let lower = html.to_ascii_lowercase();
    (lower.contains("cas/login") || lower.contains("统一身份认证") || lower.contains("请先登录"))
        && !html.contains("id=\"dataList\"")
        && !html.contains("id=\"kbtable\"")
}

fn trim_cell(text: &str) -> String {
    text.replace('\u{00a0}', " ").split_whitespace().collect::<Vec<_>>().join(" ")
}

pub fn parse_grades_from_html(html: &str) -> Result<Vec<Grade>, ParseError> {
    if looks_like_login_page(html) {
        return Err(ParseError::PageError("会话已失效，请先调用 csuft_jw_login 打开登录页".into()));
    }

    let body_text = trim_cell(&Html::parse_document(html).root_element().text().collect::<String>());
    if body_text.contains("未教评") || (body_text.contains("教学评价") && body_text.contains("未完成")) {
        return Err(ParseError::Unevaluated(
            "还有未完成的教学评价，请先完成教评后再查询成绩".into(),
        ));
    }

    let document = Html::parse_document(html);
    let table_sel = Selector::parse("#dataList").unwrap();
    let tr_sel = Selector::parse("tr").unwrap();
    let td_sel = Selector::parse("td").unwrap();
    let table = document
        .select(&table_sel)
        .next()
        .ok_or_else(|| ParseError::MissingTable("未找到成绩数据表".into()))?;

    let mut grades = Vec::new();
    for tr in table.select(&tr_sel) {
        let tds: Vec<String> = tr.select(&td_sel).map(|td| trim_cell(&td.text().collect::<String>())).collect();
        if tds.len() < 13 {
            continue;
        }
        if tds[0] == "序号" || tds[2] == "课程代码" {
            continue;
        }
        let subject = tds[3].clone();
        let score = tds[4].clone();
        if subject.is_empty() && score.is_empty() {
            continue;
        }
        let exam_kind = &tds[10];
        let status = if exam_kind == "正常考试" || exam_kind.contains('重') {
            0
        } else {
            1
        };
        grades.push(Grade {
            serial_no: tds[0].clone(),
            term: tds[1].clone(),
            code: tds[2].clone(),
            subject,
            score,
            credit: parse_float(&tds[5]),
            gpa: parse_float(&tds[7]),
            status,
            property: tds[11].clone(),
            flag: tds[8].clone(),
        });
    }

    if grades.is_empty() {
        return Err(ParseError::Unevaluated(
            "未查询到成绩数据，可能还有未完成的教学评价".into(),
        ));
    }
    Ok(grades)
}

pub fn parse_level_grades_from_html(html: &str) -> Result<Vec<LevelGrade>, ParseError> {
    if looks_like_login_page(html) {
        return Err(ParseError::PageError("会话已失效，请先调用 csuft_jw_login 打开登录页".into()));
    }
    let document = Html::parse_document(html);
    let table_sel = Selector::parse("#dataList").unwrap();
    let tr_sel = Selector::parse("tr").unwrap();
    let td_sel = Selector::parse("td").unwrap();
    let table = document
        .select(&table_sel)
        .next()
        .ok_or_else(|| ParseError::MissingTable("未找到等级考试数据表".into()))?;

    let mut items = Vec::new();
    for tr in table.select(&tr_sel) {
        let tds: Vec<String> = tr.select(&td_sel).map(|td| trim_cell(&td.text().collect::<String>())).collect();
        if tds.len() < 9 {
            continue;
        }
        if tds[0] == "序号" || tds[1].contains("考试") && tds[1].contains("名称") {
            continue;
        }
        let level = if tds[4].is_empty() { tds[7].clone() } else { tds[4].clone() };
        items.push(LevelGrade {
            no: tds[0].clone(),
            course_name: tds[1].clone(),
            level_grade: level,
            time: tds[8].clone(),
        });
    }
    Ok(items)
}

pub fn parse_courses_from_html(html: &str, request_week: i32) -> Result<WeekSchedule, ParseError> {
    if looks_like_login_page(html) {
        return Err(ParseError::PageError("会话已失效，请先调用 csuft_jw_login 打开登录页".into()));
    }
    let document = Html::parse_document(html);
    let title_sel = Selector::parse("title").unwrap();
    let title = document
        .select(&title_sel)
        .next()
        .map(|n| trim_cell(&n.text().collect::<String>()))
        .unwrap_or_default();
    if title != "学期理论课表" && !html.contains("id=\"kbtable\"") {
        return Err(ParseError::PageError("课表页面错误或未登录教务系统".into()));
    }

    let mut weekno = request_week;
    let opt_sel = Selector::parse("select#zc option[selected]").unwrap();
    if let Some(opt) = document.select(&opt_sel).next() {
        if let Some(val) = opt.value().attr("value") {
            if let Ok(v) = val.trim().parse::<i32>() {
                weekno = v;
            }
        }
    }

    let mut days: Vec<DaySchedule> = (1..=7)
        .map(|weekday| DaySchedule {
            weekday,
            courses: Vec::new(),
        })
        .collect();

    let tr_sel = Selector::parse("#kbtable tr").unwrap();
    let th_sel = Selector::parse("th").unwrap();
    let td_sel = Selector::parse("td").unwrap();
    let kb_sel = Selector::parse("div.kbcontent").unwrap();
    let font_sel = Selector::parse("font").unwrap();

    for (i, tr) in document.select(&tr_sel).enumerate() {
        if i == 0 {
            continue;
        }
        let th_text = tr
            .select(&th_sel)
            .next()
            .map(|th| trim_cell(&th.text().collect::<String>()))
            .unwrap_or_default();
        if th_text.is_empty() || th_text.starts_with("备注") {
            continue;
        }
        let (start_p, end_p) = parse_period_range(&th_text);
        if start_p == 0 && end_p == 0 {
            continue;
        }
        for (col, td) in tr.select(&td_sel).enumerate() {
            let weekday = (col as i32) + 1;
            for div in td.select(&kb_sel) {
                let name = extract_course_name(div);
                if name.is_empty() || name == "&nbsp;" {
                    continue;
                }
                let mut teacher = String::new();
                let mut classroom = String::new();
                let mut weeks_str = String::new();
                for font in div.select(&font_sel) {
                    let title = font.value().attr("title").unwrap_or("");
                    let text = trim_cell(&font.text().collect::<String>());
                    if title.contains("老师") {
                        teacher = text;
                    } else if title.contains("周次") {
                        weeks_str = text;
                    } else if title.contains("教室") {
                        classroom = text;
                    }
                }
                if weekno > 0 && !weeks_str.is_empty() && !week_in_weeks(weekno, &weeks_str) {
                    continue;
                }
                if weekday >= 1 && weekday <= 7 {
                    days[(weekday - 1) as usize].courses.push(Course {
                        name,
                        teacher,
                        classroom,
                        weekday,
                        start_period: start_p,
                        end_period: end_p,
                    });
                }
            }
        }
    }

    Ok(WeekSchedule { weekno, days })
}

pub fn parse_student_info_from_html(html: &str, sid_hint: &str) -> Result<StudentInfo, ParseError> {
    let document = Html::parse_document(html);
    let td_sel = Selector::parse("td").unwrap();
    let tr_sel = Selector::parse("tr").unwrap();
    let mut info = StudentInfo::default();
    if sid_hint.len() >= 4 {
        info.grade = sid_hint[..4].to_string();
    }
    for td in document.select(&td_sel) {
        let text = trim_cell(&td.text().collect::<String>());
        if let Some(rest) = text.strip_prefix("院系：") {
            info.college = rest.to_string();
        } else if let Some(rest) = text.strip_prefix("专业：") {
            info.major = rest.to_string();
        } else if let Some(rest) = text.strip_prefix("班级：") {
            info.class = rest.to_string();
        } else if let Some(rest) = text.strip_prefix("学号：") {
            let sid = rest.trim();
            if sid.len() >= 4 {
                info.grade = sid[..4].to_string();
            }
        }
    }
    for tr in document.select(&tr_sel) {
        let tds: Vec<String> = tr.select(&td_sel).map(|td| trim_cell(&td.text().collect::<String>())).collect();
        if tds.len() >= 2 && tds[0] == "姓名" && info.name.is_empty() {
            info.name = tds[1].clone();
        }
    }
    if info.college.is_empty() && info.major.is_empty() && info.class.is_empty() && info.name.is_empty() {
        return Err(ParseError::MissingTable("未能解析到学生信息".into()));
    }
    Ok(info)
}

pub fn parse_period_range(text: &str) -> (i32, i32) {
    let re = Regex::new(r"\d+").unwrap();
    let nums: Vec<i32> = re
        .find_iter(text)
        .filter_map(|m| m.as_str().parse().ok())
        .collect();
    match nums.as_slice() {
        [] => (0, 0),
        [start] => (*start, *start),
        [start, ..] => (*start, *nums.last().unwrap()),
    }
}

pub fn week_in_weeks(week_no: i32, weeks_str: &str) -> bool {
    let mut s = weeks_str;
    if let Some(idx) = s.find('(') {
        s = &s[..idx];
    }
    let s = s.trim();
    if s.is_empty() {
        return true;
    }
    for part in s.split(',') {
        let part = part.trim();
        if part.is_empty() {
            continue;
        }
        if let Some((a, b)) = part.split_once('-') {
            if let (Ok(start), Ok(end)) = (a.trim().parse::<i32>(), b.trim().parse::<i32>()) {
                if week_no >= start && week_no <= end {
                    return true;
                }
            }
        } else if let Ok(n) = part.parse::<i32>() {
            if week_no == n {
                return true;
            }
        }
    }
    false
}

fn extract_course_name(div: scraper::ElementRef<'_>) -> String {
    for child in div.children() {
        if let Some(text) = child.value().as_text() {
            let t = trim_cell(text);
            if !t.is_empty() {
                return t;
            }
        }
        if child.value().as_element().is_some_and(|el| el.name() == "br") {
            break;
        }
    }
    String::new()
}

fn parse_float(s: &str) -> f64 {
    s.replace(',', "").trim().parse().unwrap_or(0.0)
}

pub fn filter_grades_by_term<'a>(grades: &'a [Grade], term: &str) -> Vec<Grade> {
    if term.is_empty() {
        return grades.to_vec();
    }
    grades.iter().filter(|g| g.term == term).cloned().collect()
}
