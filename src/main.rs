extern crate native_tls;
#[macro_use]
extern crate serde;
extern crate serde_yaml;
#[macro_use]
extern crate clap;

#[derive(Deserialize)]
pub struct Account {
    host: String,
    port: u16,
    username: String,
    password: String,
    rules: Vec<Rule>,
}

#[derive(Clone, Deserialize)]
pub struct Rule {
    name: String,
    enabled: bool,
    trigger: Trigger,
    action: Action,
}

#[derive(Clone, Deserialize)]
pub enum Trigger {
    SubjectContains(String),
    SubjectStartsWith(String),
    SubjectEndsWith(String),
    SubjectExact(String),
}

#[derive(Clone, Deserialize)]
pub enum Action {
    MoveIntoMailbox(String),
}

fn main() -> Result<(), Box<dyn std::error::Error>> {
    let app = clap::App::new("mail-organizer")
        .version(crate_version!())
        .author(crate_authors!())
        .about(crate_description!())
        .arg(clap::Arg::new("configs")
            .index(1)
            .multiple(true)
        )
        .get_matches();

    let configs = app.values_of("configs").unwrap();

    for c in configs {
        if let Ok(file) = std::fs::read_to_string(&c) {
            let account: Account = serde_yaml::from_str(file.as_str())?;
            if let Ok(successful) = run(account) {
                println!("Number of processed mails for config '{}': {}", &c,  successful)
            }
        } else {
            eprintln!("Error reading file: {}", c)
        }
    }
    Ok(())
}

fn run(config: Account) -> Result<usize, imap::Error> {
    let domain: &str = config.host.as_str();

    let tls = native_tls::TlsConnector::builder().build().unwrap();

    // we pass in the domain twice to check that the server's TLS
    // certificate is valid for the domain we're connecting to.
    let client = imap::connect_starttls((domain, config.port), domain, &tls).unwrap();

    // the client we have here is unauthenticated.
    // to do anything useful with the e-mails, we need to log in
    let mut imap_session = client
        .login(config.username.as_str(), config.password.as_str())
        .map_err(|e| e.0)?;

    // Check for capabilities
    let capabilities = imap_session.capabilities()?;
    if !capabilities.has_str("UIDPLUS") {
        panic!("Server '{}', does not support UIDPLUS", config.host);
    }

    // we want to fetch the first email in the INBOX mailbox
    let mailbox = imap_session.select("INBOX")?;
    let no_msgs = mailbox.exists;

    let seq: String = "1:".to_owned() + &no_msgs.to_string();
    let messages = imap_session.fetch(seq, "(UID ENVELOPE)")?;
    let msg_iter = messages.iter();

    let result: Vec<Result<(), imap::Error>> = msg_iter
        .map(|m| -> Option<(u32, Rule)> {
            if let Some(e) = m.envelope() {
                let matched_rules: Vec<Rule> = config.rules
                    .iter()
                    .filter(|&r| -> bool {
                        if r.enabled {
                            let subject = std::str::from_utf8(&e.subject.unwrap()).unwrap();

                            match &r.trigger {
                                Trigger::SubjectContains(s) => {
                                    subject.contains(s)
                                }
                                Trigger::SubjectStartsWith(s) => {
                                    subject.starts_with(s)
                                }
                                Trigger::SubjectEndsWith(s) => {
                                    subject.ends_with(s)
                                }
                                Trigger::SubjectExact(s) => {
                                    subject == s.as_str()
                                }
                            }
                        } else {
                            false
                        }
                    })
                    .cloned()
                    .collect();

                if let Some(first_matched_rule) = matched_rules.first() {
                    Some((m.uid.unwrap().clone(), first_matched_rule.clone()))
                } else {
                    None
                }
            } else {
                None
            }
        })
        .filter_map(|m| m)
        .map(|m: (u32, Rule)| -> Result<(), imap::Error> {
            match m.1.action {
                Action::MoveIntoMailbox(target) => {
                    imap_session.uid_mv(format!("{}", m.0), target)
                }
            }
        })
        .collect();

    imap_session.logout()?;

    if result.iter().any(|r| r.is_err()) {
        Ok(0)
    } else {
        Ok(result.len())
    }
}