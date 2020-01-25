use structopt::StructOpt;

use crate::{secret::SubcommandOption, GlobalOption, SubcommandError};

use orbital_headers::orbital_types::SecretType;
use orbital_headers::secret::{secret_service_client::SecretServiceClient, SecretAddRequest};
use orbital_services::ORB_DEFAULT_URI;
use tonic::Request;

use log::debug;

use orbital_database::postgres::secret::Secret;
use prettytable::{cell, format, row, Table};

use std::fs::File;
use std::io::prelude::*;

#[derive(Debug, StructOpt, Clone)]
#[structopt(rename_all = "kebab_case")]
pub struct ActionOption {
    /// Secret name
    #[structopt(required = true)]
    secret_name: String,

    /// Secret Type
    #[structopt(long, required = true, possible_values = &SecretType::variants())]
    secret_type: SecretType,

    /// Name of Orbital org
    #[structopt(long, env = "ORB_DEFAULT_ORG")]
    org: Option<String>,

    /// Secret filepath
    #[structopt(long, short = "f")]
    secret_file: Option<String>,
}

pub async fn action_handler(
    _global_option: GlobalOption,
    _subcommand_option: SubcommandOption,
    action_option: ActionOption,
) -> Result<(), SubcommandError> {
    let mut file = File::open(&action_option.secret_file.expect("No secret filepath given"))?;
    let mut contents = String::new();
    file.read_to_string(&mut contents)?;

    let mut client = SecretServiceClient::connect(format!("http://{}", ORB_DEFAULT_URI)).await?;

    let request = Request::new(SecretAddRequest {
        org: action_option.org.unwrap_or_default().into(),
        name: action_option.secret_name.into(),
        secret_type: action_option.secret_type.into(),
        data: contents.into(),
    });

    debug!("Request for secret add: {:?}", &request);

    let response = client.secret_add(request).await;

    // By default, format the response into a table
    let mut table = Table::new();
    table.set_format(*format::consts::FORMAT_NO_BORDER_LINE_SEPARATOR);

    // Print the header row
    table.set_titles(row![
        "Org Name",
        "Secret Name",
        "Secret Type",
        "Vault Path",
        "Active State",
    ]);

    match response {
        Err(_e) => {
            eprintln!("Error adding Secret");
            Ok(())
        }
        Ok(s) => {
            let secret_proto = s.into_inner();

            debug!("RESPONSE = {:?}", &secret_proto);

            // By default, format the response into a table
            let mut table = Table::new();
            table.set_format(*format::consts::FORMAT_NO_BORDER_LINE_SEPARATOR);

            // Print the header row
            table.set_titles(row![
                "Org Name",
                "Secret Name",
                "Secret Type",
                "Vault Path",
                "Active State",
            ]);

            let secret = Secret::from(secret_proto.clone());

            table.add_row(row![
                secret_proto.org,
                secret.name,
                &format!("{:?}", secret.secret_type),
                secret.vault_path,
                &format!("{:?}", secret.active_state),
            ]);

            // Print the table to stdout
            table.printstd();
            Ok(())
        }
    }
}
