variable "aws_region" {
  type        = string
  default     = "us-east-1"
  description = "Región de AWS para desplegar los recursos"
}

variable "project_name" {
  type        = string
  default     = "novobanco"
  description = "Nombre base del proyecto"
}

variable "db_password" {
  type        = string
  default     = "ProdSuperSecurePassword123!"
  sensitive   = true
  description = "Contraseña maestra para la base de datos de producción RDS"
}
