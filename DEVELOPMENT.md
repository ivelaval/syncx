# Development Workflow Guide

Esta guía explica las mejores prácticas para desarrollar y versionar el proyecto SyncX.

## 📁 Directory Structure Overview

SyncX organizes all cloned repositories under a `projects/` subdirectory with automatic path cleaning:

**Example:**
- Base directory: `/Users/vennet/Olive.com`
- Projects location: `/Users/vennet/Olive.com/projects/`
- Example project: `/Users/vennet/Olive.com/projects/analytics/fenske/`

The application automatically removes redundant organizational prefixes (`uproarcar`, `olive-com`) from Git URLs to keep paths clean.

## 🚀 Flujo de Trabajo de Desarrollo Rápido

Cuando estás desarrollando y haciendo cambios frecuentes:

### Opción 1: Instalación Rápida (Recomendado)

```bash
# 1. Edita tu código
vim internal/git.go

# 2. Instala rápidamente (solo compila tu plataforma)
make install-dev

# 3. Prueba inmediatamente
syncx --version
syncx clone --file projects.json ...
```

**Ventajas:**
- ⚡ Muy rápido (solo compila para tu plataforma)
- 🔄 Actualiza el binario instalado automáticamente
- 🏷️ Marca la versión como `-dev` para distinguirla

### Opción 2: Compilación Local

```bash
# Compila sin instalar
make build-dev

# Ejecuta directamente
./syncx --version
./syncx clone --file projects.json ...
```

### Opción 3: Ejecución Directa (Sin Compilar)

```bash
# Ejecuta sin compilar (más lento pero útil para debugging)
go run main.go --version
go run main.go clone --file projects.json ...
```

## 📦 Flujo de Trabajo de Producción

Cuando estás listo para crear una release oficial:

### 1. Verificar Estado

```bash
# Ver versión actual
make version

# Output:
# Current version: 2.1.0
# Git commit: abc123
# Git branch: main
```

### 2. Actualizar Versión

```bash
# Para bug fixes (2.1.0 -> 2.1.1)
make bump-patch

# Para nuevas features (2.1.0 -> 2.2.0)
make bump-minor

# Para cambios importantes (2.1.0 -> 3.0.0)
make bump-major
```

### 3. Confirmar y Etiquetar

```bash
# Commit del cambio de versión
git add VERSION
git commit -m "Bump version to v2.2.0"

# Crear tag
git tag -a v2.2.0 -m "Release v2.2.0

Features:
- Added empty repository detection
- Improved error reporting
- ..."

# Push con tags
git push origin main
git push origin v2.2.0
```

### 4. Compilar para Producción

```bash
# Compila para todas las plataformas
make build

# Output:
# - build/syncx (tu plataforma)
# - build/syncx-darwin-amd64
# - build/syncx-darwin-arm64
# - build/syncx-linux-amd64
# - build/syncx-linux-arm64
# - build/syncx-windows-amd64.exe
```

### 5. Instalar Localmente

```bash
# Instala la versión de producción
make install

# Verifica
syncx --version
# Output: 2.2.0 (built: 2025-01-15_10:30:00, commit: abc1234)
```

## 🔄 Comparación de Comandos

| Comando | Uso | Velocidad | Cuándo Usar |
|---------|-----|-----------|-------------|
| `make install-dev` | Desarrollo diario | ⚡⚡⚡ Muy rápido | Cambios frecuentes, testing |
| `make build-dev` | Compilar localmente | ⚡⚡ Rápido | Testing sin instalar |
| `go run main.go` | Ejecutar sin compilar | ⚡ Normal | Debugging rápido |
| `make build` | Compilar producción | 🐢 Lento | Releases oficiales |
| `make install` | Instalar producción | 🐢 Lento | Instalar release |

## 📋 Versionamiento Semántico

Seguimos [Semantic Versioning 2.0.0](https://semver.org/):

```
MAJOR.MINOR.PATCH
  │     │      │
  │     │      └─ Bug fixes (2.1.0 -> 2.1.1)
  │     └──────── Nuevas features compatibles (2.1.0 -> 2.2.0)
  └────────────── Cambios incompatibles (2.1.0 -> 3.0.0)
```

### Ejemplos:

- **Patch (2.1.0 → 2.1.1)**: Bug fix, corrección de typo, performance
- **Minor (2.1.0 → 2.2.0)**: Nueva funcionalidad, nueva opción CLI
- **Major (2.1.0 → 3.0.0)**: Cambio en API, remover features deprecadas

## 🛠️ Comandos Útiles

```bash
# Ver ayuda completa
make help

# Ver versión actual
make version

# Formatear código
make fmt

# Ejecutar tests
make test

# Limpiar binarios
make clean

# Desinstalar
make uninstall
```

## 📝 Ejemplo Completo: Agregar Nueva Feature

```bash
# 1. Crear branch para la feature
git checkout -b feature/empty-repo-detection

# 2. Desarrollar y probar iterativamente
vim internal/git.go
make install-dev
syncx --version  # Verás: 2.1.0-dev
syncx clone ...  # Probar

# 3. Cuando esté listo
git add .
git commit -m "Add empty repository detection"

# 4. Merge a main
git checkout main
git merge feature/empty-repo-detection

# 5. Bump version (nueva feature = minor)
make bump-minor  # 2.1.0 -> 2.2.0

# 6. Tag y release
git commit -am "Bump version to v2.2.0"
git tag -a v2.2.0 -m "Release v2.2.0"

# 7. Build production
make build

# 8. Install local
make install

# 9. Verify
syncx --version  # Verás: 2.2.0 (built: ...)

# 10. Push
git push origin main --tags
```

## 🎯 Respuesta a tu Pregunta Original

**P: ¿Cómo genero un nuevo ejecutable cuando hago cambios?**

**R:**

### Durante Desarrollo (cambios frecuentes):
```bash
make install-dev  # ⚡ Rápido, solo tu plataforma
```

### Para Release (versión oficial):
```bash
make bump-minor   # Actualizar versión
make build        # Compilar todas las plataformas
make install      # Instalar localmente
```

## 💡 Tips

1. **Usa `make install-dev`** para desarrollo diario - es mucho más rápido
2. **Usa `make build`** solo para releases oficiales
3. **No commits el binario** - está en `.gitignore`
4. **Siempre bump version** antes de hacer un release
5. **Usa tags de git** para releases oficiales

## 🔍 Verificar Información de Build

```bash
# Ver información completa
syncx --version

# Output con versión de desarrollo:
# 2.1.0-dev (built: 2025-01-15_10:30:00, commit: abc1234)

# Output con versión de producción:
# 2.1.0 (built: 2025-01-15_10:30:00, commit: abc1234)
```

## 📚 Más Información

- [README.md](README.md) - Información general del proyecto
- [INSTALL.md](INSTALL.md) - Guía de instalación para usuarios
- [CLAUDE.md](CLAUDE.md) - Documentación de comandos
