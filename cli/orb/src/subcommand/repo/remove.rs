use structopt::StructOpt;

use crate::{repo::SubcommandOption, GlobalOption, SubcommandError};

use orbital_headers::code::{code_service_client::CodeServiceClient, GitRepoRemoveRequest};
use orbital_services::ORB_DEFAULT_URI;
use tonic::Request;

use git_meta::git_info;
use log::debug;
use std::path::PathBuf;

#[derive(Debug, StructOpt, Clone)]
#[structopt(rename_all = "kebab_case")]
pub struct ActionOption {
    /// Repo path
    #[structopt(parse(from_os_str), env = "PWD")]
    path: PathBuf,

    /// Name of Orbital org
    #[structopt(long, env = "ORB_DEFAULT_ORG")]
    org: Option<String>,
}

pub async fn action_handler(
    _global_option: GlobalOption,
    _subcommand_option: SubcommandOption,
    action_option: ActionOption,
) -> Result<(), SubcommandError> {
    let repo_info =
        match git_info::get_git_info_from_path(&action_option.path.as_path(), &None, &None) {
            Ok(info) => info,
            Err(_e) => panic!("Unable to parse path for git repo info"),
        };

    let request = Request::new(GitRepoRemoveRequest {
        org: action_option.org.unwrap_or_default(),
        git_provider: repo_info.provider,
        name: repo_info.repo,
        //user: ,
        uri: repo_info.uri,
        //force: ,
        ..Default::default()
    });

    debug!("Request for git repo remove: {:?}", &request);

    let mut client = CodeServiceClient::connect(format!("http://{}", ORB_DEFAULT_URI)).await?;

    let response = client.git_repo_remove(request).await?;
    println!("RESPONSE = {:?}", response);
    Ok(())
}
