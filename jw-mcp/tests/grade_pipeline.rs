use csuft_jw_mcp::gpa::{calculate_gpa, credit_summary};
use csuft_jw_mcp::mcp::{decode_message, encode_message, handle_message};
use csuft_jw_mcp::parse::{
    filter_grades_by_term, parse_courses_from_html, parse_grades_from_html, parse_level_grades_from_html,
    parse_student_info_from_html, week_in_weeks,
};
use csuft_jw_mcp::query::build_grades_payload;
use serde_json::json;

fn fixture(name: &str) -> String {
    let path = format!("{}/tests/fixtures/{name}", env!("CARGO_MANIFEST_DIR"));
    std::fs::read_to_string(path).expect("fixture")
}

#[test]
fn parse_grades_reads_real_table_cells() {
    let html = fixture("grades.html");
    let grades = parse_grades_from_html(&html).expect("parse grades");
    let math = grades.iter().find(|g| g.code == "B0001").expect("math");
    assert_eq!(math.subject, "高等数学A1");
    assert_eq!(math.score, "90");
    assert_eq!(math.credit, 5.0);
    assert_eq!(math.property, "必修");
    assert_eq!(math.status, 0);

    let makeup = grades.iter().find(|g| g.code == "B0003" && g.status == 1).expect("makeup");
    assert_eq!(makeup.score, "72");
    assert_eq!(makeup.term, "2023-2024-2");

    let deferred = grades.iter().find(|g| g.flag == "缓考").expect("deferred");
    assert_eq!(deferred.subject, "操作系统");
}

#[test]
fn gpa_uses_required_courses_and_skips_deferred() {
    let html = fixture("grades.html");
    let grades = parse_grades_from_html(&html).unwrap();
    let makeup = grades.iter().find(|g| g.code == "B0003" && g.status == 1).unwrap();
    assert_eq!(makeup.score, "72");
    let deferred = grades.iter().find(|g| g.flag == "缓考").unwrap();
    assert_eq!(deferred.subject, "操作系统");

    let without_deferred: Vec<_> = grades.iter().filter(|g| g.flag != "缓考").cloned().collect();
    let full = calculate_gpa(&grades);
    let no_deferred = calculate_gpa(&without_deferred);
    assert_eq!(full, no_deferred, "缓考 must not change GPA");

    let required_only: Vec<_> = grades.iter().filter(|g| g.property == "必修").cloned().collect();
    let with_elective = calculate_gpa(&grades);
    let required_gpa = calculate_gpa(&required_only);
    assert_eq!(with_elective, required_gpa, "选修 must not change required GPA");

    let without_makeup: Vec<_> = grades
        .iter()
        .filter(|g| !(g.code == "B0003" && g.status == 1))
        .cloned()
        .collect();
    let with_makeup = calculate_gpa(&grades);
    let no_makeup = calculate_gpa(&without_makeup);
    assert_ne!(with_makeup.average_gpa, no_makeup.average_gpa, "补考 must change GPA");

    let credits = credit_summary(&grades);
    assert!(credits.required_credits >= 5.0 + 3.0 + 2.0);
    assert!(credits.earned_required_credits >= 5.0);
    let elective = grades.iter().find(|g| g.property == "选修").unwrap();
    assert_eq!(elective.subject, "摄影基础");
}

#[test]
fn term_filter_and_summary_use_parsed_grades() {
    let html = fixture("grades.html");
    let all = parse_grades_from_html(&html).unwrap();
    let term = filter_grades_by_term(&all, "2023-2024-1");
    assert!(term.iter().all(|g| g.term == "2023-2024-1"));
    assert_eq!(term.len(), 2);
    let payload = build_grades_payload("2023-2024-1", term, &all);
    assert_eq!(payload.count, 2);
    assert!(payload.by_term.contains_key("2023-2024-1"));
    assert!(payload.by_term.contains_key("2023-2024-2"));
}

#[test]
fn parse_level_exams_prefers_numeric_score() {
    let html = fixture("level.html");
    let items = parse_level_grades_from_html(&html).unwrap();
    assert_eq!(items[0].course_name, "全国大学英语四级考试");
    assert_eq!(items[0].level_grade, "498");
    assert_eq!(items[1].level_grade, "合格");
    assert_eq!(items[1].time, "2024-12-01");
}

#[test]
fn parse_courses_filters_week_from_kbcontent() {
    let html = fixture("courses.html");
    let week3 = parse_courses_from_html(&html, 3).unwrap();
    assert_eq!(week3.weekno, 3);
    let monday = &week3.days[0].courses;
    assert!(monday.iter().any(|c| c.name == "操作系统" && c.classroom == "云塘A101"));
    assert!(!monday.iter().any(|c| c.name == "形势与政策"));
    let wednesday = &week3.days[2].courses;
    assert!(wednesday.iter().any(|c| c.name == "大学体育"));

    assert!(!week3.days[0].courses.iter().any(|c| c.name == "形势与政策"));
    assert!(week_in_weeks(8, "8-9(周)"));
    assert!(!week_in_weeks(3, "8-9(周)"));
}

#[test]
fn week_range_helper_matches_schedule_strings() {
    assert!(week_in_weeks(3, "1-16(周)"));
    assert!(week_in_weeks(7, "3,5,7(周)"));
    assert!(!week_in_weeks(4, "3,5,7(周)"));
    assert!(!week_in_weeks(10, "8-9(周)"));
}

#[test]
fn parse_student_info_from_grade_page() {
    let html = fixture("grades.html");
    let info = parse_student_info_from_html(&html, "").unwrap();
    assert_eq!(info.college, "计算机与数学学院");
    assert_eq!(info.major, "计算机科学与技术");
    assert_eq!(info.class, "2023计算机科学与技术2班");
    assert_eq!(info.name, "测试同学");
    assert_eq!(info.grade, "2023");
}

#[tokio::test]
async fn mcp_initialize_and_lists_grade_tools() {
    let init = handle_message(json!({
        "jsonrpc": "2.0",
        "id": 1,
        "method": "initialize",
        "params": { "protocolVersion": "2024-11-05", "capabilities": {}, "clientInfo": { "name": "test", "version": "0" } }
    }))
    .await;
    assert_eq!(init["result"]["serverInfo"]["name"], "csuft-jw-mcp");
    assert!(init["result"]["capabilities"]["tools"].is_object());

    let listed = handle_message(json!({
        "jsonrpc": "2.0",
        "id": 2,
        "method": "tools/list",
        "params": {}
    }))
    .await;
    let names: Vec<&str> = listed["result"]["tools"]
        .as_array()
        .unwrap()
        .iter()
        .map(|t| t["name"].as_str().unwrap())
        .collect();
    for required in [
        "csuft_jw_login",
        "csuft_jw_open_login_page",
        "csuft_jw_get_grades",
        "csuft_jw_get_gpa",
        "csuft_jw_get_level_exams",
        "csuft_jw_get_courses",
    ] {
        assert!(names.contains(&required), "missing {required}");
    }
    let blob = listed.to_string();
    for forbidden in ["评教", "改密", "改密码", "改绑定", "unbind", "evaluate"] {
        assert!(!blob.contains(forbidden), "write tool leaked: {forbidden}");
    }
    let descs: String = listed["result"]["tools"]
        .as_array()
        .unwrap()
        .iter()
        .map(|t| format!("{} {}", t["name"], t["description"]))
        .collect::<Vec<_>>()
        .join("\n");
    assert!(descs.contains("成绩"));
    assert!(descs.contains("课表"));
    assert!(descs.contains("绩点") || descs.contains("学分"));
    assert!(descs.contains("等级考试"));
    assert!(descs.contains("登录"));
}

#[test]
fn login_endpoints_open_cas_and_redirect_to_jwgl() {
    let campus = csuft_jw_mcp::session::Endpoints::from_mode("campus");
    let webvpn = csuft_jw_mcp::session::Endpoints::from_mode("webvpn");
    let default = csuft_jw_mcp::session::Endpoints::from_mode("");
    assert_eq!(default.mode, "webvpn");
    assert!(default.login_url.contains("webvpn.csuft.edu.cn"));
    for ep in [&campus, &webvpn] {
        assert!(ep.login_url.contains("/cas/login"), "{}", ep.login_url);
        assert!(
            ep.login_url.contains("jwgl") || ep.main_url.contains("jwgl") || ep.main_url.contains("jsxsd"),
            "missing 教务 redirect {}",
            ep.login_url
        );
        assert!(ep.grade_url.contains("kscj"));
        assert!(ep.level_url.contains("djkscj") || ep.level_url.contains("kscj"));
        assert!(ep.course_url.contains("xskb"));
    }
}

#[tokio::test]
async fn mcp_query_tools_without_session_refuse_and_do_not_write() {
    let dir = std::env::temp_dir().join(format!("csuft-jw-mcp-test-{}", std::process::id()));
    std::fs::create_dir_all(&dir).unwrap();
    std::env::set_var("CSUFT_JW_MCP_DIR", &dir);
    let _ = csuft_jw_mcp::session::clear_session();
    let calls = [
        ("csuft_jw_get_grades", json!({})),
        ("csuft_jw_get_gpa", json!({})),
        ("csuft_jw_get_level_exams", json!({})),
        ("csuft_jw_get_courses", json!({"term": "2025-2026-1", "week": 1})),
        ("csuft_jw_get_student_info", json!({})),
    ];
    for (i, (name, arguments)) in calls.iter().enumerate() {
        let resp = handle_message(json!({
            "jsonrpc": "2.0",
            "id": i + 10,
            "method": "tools/call",
            "params": { "name": name, "arguments": arguments }
        }))
        .await;
        assert_eq!(resp["result"]["isError"], true, "{name} should error");
        let text = resp["result"]["content"][0]["text"].as_str().unwrap();
        assert!(
            text.contains("尚未登录") || text.contains("csuft_jw_login"),
            "{name}: {text}"
        );
        assert!(!text.contains("评教") && !text.contains("改密"));
    }
}

#[test]
fn mcp_framing_roundtrip_uses_content_length() {
    let msg = json!({"jsonrpc":"2.0","id":9,"method":"ping"});
    let encoded = encode_message(&msg);
    let header = std::str::from_utf8(&encoded).unwrap();
    assert!(header.starts_with("Content-Length:"));
    let decoded = decode_message(&encoded).unwrap();
    assert_eq!(decoded["method"], "ping");
    assert_eq!(decoded["id"], 9);
}
