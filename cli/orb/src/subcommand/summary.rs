use structopt::StructOpt;

use crate::GlobalOption;

use orbital_headers::build_meta::{
    build_service_client::BuildServiceClient, BuildSummaryRequest, BuildTarget,
};

use orbital_database::postgres::schema::JobState;

use anyhow::Result;
use chrono::NaiveDateTime;
use git_meta::git_info;
use log::debug;
use orbital_services::ORB_DEFAULT_URI;
use prettytable::{cell, format, row, Table};
use std::path::PathBuf;
use tonic::Request;

/// Local options for customizing summary request
#[derive(Debug, StructOpt)]
#[structopt(rename_all = "kebab_case")]
pub struct SubcommandOption {
    /// Limit number of results
    #[structopt(long, default_value = "5")]
    limit: i32,

    /// Git commit hash (Default is to choose the remote HEAD commit)
    #[structopt(long)]
    hash: Option<String>,

    /// Branch name (Default is to choose checked out branch)
    #[structopt(long)]
    branch: Option<String>,

    /// Path to local repo. Defaults to current working directory
    #[structopt(long, parse(from_os_str), env = "PWD")]
    path: PathBuf,

    /// Name of Orbital org
    #[structopt(long, env = "ORB_DEFAULT_ORG")]
    org: Option<String>,

    /// Print full commit hash
    #[structopt(long, short)]
    wide: bool,
}

// FIXME: Request for summary is not currently served well by proto. How to differeniate from a regular log request?
// Idea: Need a get_summary call. Build id should be Option<u32>, so we can summarize a repo or a specific build
/// *Not yet implemented* Generates request for build summaries
pub async fn subcommand_handler(
    _global_option: GlobalOption,
    local_option: SubcommandOption,
) -> Result<()> {
    let mut client = BuildServiceClient::connect(format!("http://{}", ORB_DEFAULT_URI)).await?;

    // Path
    let path = &local_option.path;

    let git_context =
        git_info::get_git_info_from_path(path, &local_option.branch, &local_option.hash)?;

    // Idea: Index should be Option<u32>
    let request = Request::new(BuildSummaryRequest {
        build: Some(BuildTarget {
            org: local_option.org.expect("Please provide an org name"),
            git_repo: git_context.git_url.clone().name,
            remote_uri: git_context.git_url.trim_auth().to_string(),
            branch: git_context.branch,
            commit_hash: git_context.commit_id,
            ..Default::default()
        }),
        limit: local_option.limit,
    });

    let response = client.build_summary(request).await?.into_inner();

    // By default, format the response into a table
    let mut table = Table::new();
    table.set_format(*format::consts::FORMAT_NO_BORDER_LINE_SEPARATOR);

    // Print the header row
    table.set_titles(row![
        bc =>
        "Build #",
        "Org",
        "Repo",
        "Branch",
        "Commit",
        "User Envs",
        "Queue time",
        "Start time",
        "End time",
        "Build state",
    ]);

    //println!("RESPONSE = {:?}", response);
    match response.summaries.len() {
        0 => {
            println!("No summaries found");
        }
        _ => {
            for summary in &response.summaries {
                let build_target = summary.build.clone().expect("No build target in summary");

                let commit = match &local_option.wide {
                    true => build_target.commit_hash,
                    false => build_target.commit_hash[..7].to_string(),
                };

                //queue_time: NaiveDateTime::from_timestamp(Utc::now().timestamp(), 0),

                let queue_time = match &summary.queue_time {
                    Some(t) => format!(
                        "{:?}",
                        NaiveDateTime::from_timestamp(t.seconds, t.nanos as u32)
                    ),
                    None => format!("---"),
                };

                let start_time = match &summary.start_time {
                    Some(t) => format!(
                        "{:?}",
                        NaiveDateTime::from_timestamp(t.seconds, t.nanos as u32)
                    ),
                    None => format!("---"),
                };

                let end_time = match &summary.end_time {
                    Some(t) => format!(
                        "{:?}",
                        NaiveDateTime::from_timestamp(t.seconds, t.nanos as u32)
                    ),
                    None => format!("---"),
                };

                table.add_row(row![
                    build_target.id,
                    build_target.org,
                    build_target.git_repo,
                    build_target.branch,
                    commit,
                    build_target.user_envs,
                    queue_time,
                    start_time,
                    end_time,
                    JobState::from(summary.build_state),
                ]);
            }

            debug!("RESPONSE = {:?}", &response);
        }
    }

    // Print the table to stdout
    table.printstd();

    Ok(())
}
