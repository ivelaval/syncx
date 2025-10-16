# 📂 Output Directory Guide - Smart Repository Organization

El Olive Clone Assistant v2.0 ahora incluye un sistema inteligente para manejar directorios de salida, evitando que los repositorios clonados queden mezclados con los archivos del script.

## 🎯 **Nueva Funcionalidad: `--output` / `-o`**

### **Comando Mejorado**
```bash
# Nueva opción --output (recomendada)
./olive-clone clone --output /Users/vennet/projects

# Forma corta
./olive-clone clone -o ~/repositories

# Rutas relativas también funcionan
./olive-clone clone -o ../my-repos
```

### **Directorio por Defecto Inteligente**
Si no especificas `--output`, el sistema usa automáticamente:
- **`../repositories`** (fuera de la carpeta del script)
- **`~/repositories`** (si está instalado en sistema)
- **`./repositories`** (como fallback)

## 📍 **Ejemplos de Uso**

### **Rutas Absolutas**
```bash
# Directorio específico en tu home
./olive-clone clone -o /Users/vennet/projects

# En cualquier ubicación del sistema
./olive-clone clone -o /opt/repositories
```

### **Rutas Relativas**
```bash
# Un nivel arriba del script
./olive-clone clone -o ../repositories

# En el directorio actual (subcarpeta)
./olive-clone clone -o ./local-repos

# Usando ~ para home directory
./olive-clone clone -o ~/Projects
```

### **Con Otras Opciones**
```bash
# Combinando con otras funcionalidades
./olive-clone clone -o ~/projects --group Frontend --parallel 3

# Con dry-run para probar
./olive-clone clone -o /tmp/test --dry-run --verbose
```

## 🧙‍♂️ **Wizard Interactivo Mejorado**

El wizard ahora incluye selección inteligente de directorio:

```bash
./olive-clone wizard
```

**Nueva sección de selección:**
```
📂 Output Directory Selection
Choose where to clone the repositories:

▶ /Users/vennet/repositories (Smart default - outside script folder)
  ../repositories (Parent directory) 
  /Users/vennet/repositories (Home directory)
  /Users/vennet/Projects (Home/Projects)
  ./repositories (Current directory)
  Custom path (you'll type it)
```

### **Modo Personalizado**
Si eliges "Custom path", obtienes guía interactiva:
```
💡 Examples of valid paths:
   /Users/vennet/projects
   ~/projects
   ../my-repositories
   ./local-repos

Enter custom output directory path: _
```

## 📊 **Información de Directorio (Verbose)**

Con `--verbose`, obtienes información detallada:

```bash
./olive-clone clone -o ~/projects --verbose
```

**Salida:**
```
📂 Output Directory Information
═════════════════════════════════
   Output Path: /Users/vennet/projects
   Relative Path: ../../projects
   Status: Will be created
```

## 🔧 **Funcionalidades Inteligentes**

### **Creación Automática**
```bash
# El directorio se crea automáticamente si no existe
./olive-clone clone -o ~/new-projects
```
```
ℹ️  Creating output directory: /Users/vennet/new-projects
✅ Created output directory: /Users/vennet/new-projects
```

### **Validación de Rutas**
- ✅ Rutas absolutas y relativas
- ✅ Expansión de `~` (home directory)  
- ✅ Verificación de permisos de escritura
- ✅ Creación automática de directorios padre

### **Detección Inteligente**
El sistema detecta automáticamente:
- 🏠 **Instalación en home**: Si el ejecutable está en `~/bin`
- 🌐 **Instalación del sistema**: Si está en `/usr/local/bin` o `/usr/bin`
- 📁 **Ejecución local**: Si se ejecuta desde la carpeta del proyecto

## ⚙️ **Migración desde `--directory`**

### **Opción Deprecada**
```bash
# DEPRECADO (pero aún funciona)
./olive-clone clone --directory ./repos
```
```
⚠️  --directory is deprecated, use --output or -o instead
```

### **Nueva Sintaxis**
```bash  
# NUEVA (recomendada)
./olive-clone clone --output ./repos
./olive-clone clone -o ./repos
```

## 📋 **Casos de Uso Comunes**

### **Para Desarrolladores Individuales**
```bash
# Organizar en carpeta personal
./olive-clone wizard  # Selecciona ~/Projects en el wizard

# O directamente
./olive-clone clone -o ~/Projects
```

### **Para Equipos de Desarrollo**
```bash
# Directorio compartido del equipo
./olive-clone clone -o /shared/repositories

# Con configuración específica
./olive-clone clone -o /team/frontend --group Frontend --parallel 5
```

### **Para CI/CD y Automation**
```bash
# Directorio temporal para builds
./olive-clone clone -o /tmp/build-repos --dry-run

# Directorio específico del pipeline
./olive-clone clone -o $WORKSPACE/repositories
```

### **Para Testing y Desarrollo**
```bash
# Test con directorio temporal
./olive-clone clone -o /tmp/test-$(date +%s) --dry-run

# Desarrollo local
./olive-clone clone -o ../development-repos --group Backend
```

## 🎨 **Estructura de Salida**

Los repositorios mantienen su estructura organizacional:

```
/Users/vennet/projects/
├── olive/
│   ├── frontend/
│   │   ├── main-app/
│   │   ├── admin-dashboard/
│   │   └── mobile-app/
│   ├── backend/
│   │   ├── api-server/
│   │   ├── auth-service/
│   │   └── microservices/
│   │       ├── user-service/
│   │       └── order-service/
│   └── devops/
│       ├── docker-configs/
│       └── k8s-manifests/
```

## 🔍 **Verificación de Configuración**

### **Ver Configuración Actual**
```bash
./olive-clone clone --help
```

### **Probar con Dry Run**
```bash
./olive-clone clone -o ~/test-location --dry-run --verbose
```

### **Verificar Status**
```bash  
./olive-clone status -o ~/projects --verbose
```

## 💡 **Consejos y Mejores Prácticas**

### ✅ **Recomendado**
- Usar `--output` o `-o` en lugar de `--directory`
- Dejar que el sistema use defaults inteligentes
- Usar rutas absolutas para scripts automatizados
- Probar con `--dry-run` antes de ejecutar

### ⚠️ **Evitar**
- No especificar directorio (usa automáticamente fuera del script)
- Usar `--directory` (deprecado)
- Clonar dentro de la carpeta del script
- Rutas con espacios sin quotes en scripts

### 🏆 **Ejemplos Perfectos**
```bash
# Mejor práctica para uso personal
./olive-clone wizard  # Deja que el wizard guíe

# Mejor práctica para automation
./olive-clone clone -o "$HOME/repositories" --group Backend

# Mejor práctica para testing
./olive-clone clone -o /tmp/test-repos --dry-run
```

---

**La nueva funcionalidad de output directory hace que la gestión de repositorios sea más limpia, organizada y profesional! 🎯**