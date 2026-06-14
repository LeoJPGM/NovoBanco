resource "aws_secretsmanager_secret" "db_secret" {
  name                    = "${var.project_name}-db-credentials"
  recovery_window_in_days = 0
}

resource "aws_secretsmanager_secret_version" "db_secret_val" {
  secret_id = aws_secretsmanager_secret.db_secret.id
  secret_string = jsonencode({
    DB_HOST     = aws_db_instance.postgres.address
    DB_PORT     = "5432"
    DB_USER     = aws_db_instance.postgres.username
    DB_PASSWORD = var.db_password
    DB_NAME     = aws_db_instance.postgres.db_name
  })
}
