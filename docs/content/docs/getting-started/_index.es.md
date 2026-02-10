---
title: "Primeros Pasos"
weight: 10
type: docs
next: /docs/usage
---

Ponga en marcha **AWS Doctor** en menos de un minuto.

## Instalación

### 1. Script de una sola línea (Linux y macOS)
La forma más rápida de instalar la última versión:

```bash
curl -sSfL https://raw.githubusercontent.com/elC0mpa/aws-doctor/main/install.sh | sh
```

### 2. Usando Go
Si tiene Go instalado (1.23+):

```bash
go install github.com/elC0mpa/aws-doctor@latest
```

### 3. Descarga Manual del Binario
Descargue el binario precompilado para su arquitectura desde la página de [GitHub Releases](https://github.com/elC0mpa/aws-doctor/releases). Plataformas compatibles:
- **macOS** (Intel y Apple Silicon)
- **Linux** (amd64 y arm64)
- **Windows** (amd64)

---

## Requisitos Previos

### Credenciales de AWS
**AWS Doctor** utiliza el SDK de AWS estándar para Go. Buscará automáticamente credenciales en:
1. Variables de entorno (`AWS_ACCESS_KEY_ID`, `AWS_SECRET_ACCESS_KEY`).
2. Archivo de credenciales compartido (`~/.aws/credentials`).
3. Roles de IAM para EC2/ECS si se ejecuta dentro de AWS.

### Permisos Mínimos
La herramienta requiere acceso de **Solo Lectura** para realizar auditorías.

{{< callout type="info" >}}
**Ejecución de Riesgo Cero**: Para la experiencia más sencilla y segura, recomendamos utilizar la política gestionada de AWS **`ReadOnlyAccess`**. Esto asegura que la herramienta tenga la visibilidad necesaria en todos los servicios para ejecutar todos los flujos sin capacidades de modificación.
{{< /callout >}}

Aunque `ReadOnlyAccess` es la forma más fácil de comenzar, **AWS Doctor** también admite políticas de IAM granulares. Cada funcionalidad descrita en esta documentación (como [Almacenamiento S3](../waste-detection/storage/) o [Cómputo](../waste-detection/compute/)) incluye una sección dedicada que enumera los permisos de IAM exactos requeridos.

Como mínimo, para una funcionalidad completa, una política personalizada debe incluir:
- `ce:GetCostAndUsage`
- `ec2:Describe*`
- `s3:ListAllMyBuckets`, `s3:GetLifecycleConfiguration`, `s3:ListBucketMultipartUploads`
- `elasticloadbalancing:DescribeLoadBalancers`, `elasticloadbalancing:DescribeTargetGroups`
