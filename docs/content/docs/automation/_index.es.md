---
title: "Automatización y CI/CD"
weight: 50
type: docs
prev: /docs/waste-detection
---

**AWS Doctor** está diseñado para ser parte de un ecosistema más grande. Con soporte nativo para JSON, puede integrarlo en sus flujos de trabajo automatizados.

## Salida JSON

Para obtener datos legibles por máquina, utilice el flag `--output json`:

```bash
aws-doctor --waste --output json > report.json
```

### Ejemplo de Esquema
La salida JSON proporciona una lista estructurada de cada recurso identificado como desperdicio, incluyendo su ID, fecha de creación y tamaño.

```json
{
  "account_id": "123456789012",
  "generated_at": "2026-02-09T12:00:00Z",
  "unused_ebs_volumes": [
    {
      "volume_id": "vol-0abcd1234",
      "size": 50,
      "status": "available"
    }
  ],
  "has_waste": true
}
```

---

## Casos de Uso

### 1. Fallo de Build ante Desperdicio
En su pipeline de CI/CD (GitHub Actions, Jenkins, etc.), puede usar `jq` para hacer que el build falle si la herramienta detecta algún desperdicio:

```bash
# Lógica de ejemplo
if aws-doctor --waste --output json | jq -e '.has_waste == true'; then
  echo "¡Desperdicio detectado! Limpie antes de continuar."
  exit 1
fi
```

### 2. Tableros Personalizados
Envíe la salida JSON a un stack ELK, CloudWatch Logs o una base de datos personalizada para realizar un seguimiento de la salud de su infraestructura a lo largo del tiempo.

---

## Configuración Cero
Debido a que la herramienta se basa en las credenciales estándar de AWS, funciona directamente en entornos como:
- **GitHub Actions Runners** (usando `aws-actions/configure-aws-credentials`).
- **GitLab CI** (usando variables de runner preconfiguradas).
- **Hooks de Post-Apply de Terraform** para verificar la higiene del despliegue.
