package envname

type EnvType string

const Production EnvType = "production"
const Development EnvType = "development"
const Testing EnvType = "testint"

const Port EnvType = "port"
const Host EnvType = "host"

const DbName EnvType = "db_name"
const DbDriver EnvType = "db_driver"
const DbUser EnvType = "db_user"
const DbPassword EnvType = "db_password"
const DbPort EnvType = "db_port"
const DbHost EnvType = "db_host"

const SmtpServer EnvType = "smtp_server"
const SmtpEmail EnvType = "smtp_email"
const SmtpPassword EnvType = "smtp_password"
const SmtpHost EnvType = "smtp_host"
const SmtpIdentity EnvType = "smtp_identity"

const CryptSecretKey EnvType = "crypt_secret_key"
const AesIV EnvType = "aes_iv"
