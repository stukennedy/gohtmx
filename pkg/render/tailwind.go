package render

// TailwindConfig provides default Tailwind CSS configuration for irgo apps.
const TailwindConfig = `/** @type {import('tailwindcss').Config} */
module.exports = {
  content: [
    "./**/*.templ",
    "./**/*.go",
  ],
  theme: {
    extend: {
      // Datastar and general animations
      animation: {
        'morph-in': 'morphIn 0.3s ease-out',
        'morph-out': 'morphOut 0.3s ease-out',
        'fade-in': 'fadeIn 0.3s ease-out',
        'fade-out': 'fadeOut 0.3s ease-out',
        'slide-in': 'slideIn 0.3s ease-out',
        'slide-out': 'slideOut 0.3s ease-out',
      },
      keyframes: {
        'morphIn': {
          '0%': { opacity: '0', transform: 'translateY(-10px)' },
          '100%': { opacity: '1', transform: 'translateY(0)' },
        },
        'morphOut': {
          '0%': { opacity: '1' },
          '100%': { opacity: '0' },
        },
        'fadeIn': {
          '0%': { opacity: '0' },
          '100%': { opacity: '1' },
        },
        'fadeOut': {
          '0%': { opacity: '1' },
          '100%': { opacity: '0' },
        },
        'slideIn': {
          '0%': { opacity: '0', transform: 'translateX(-10px)' },
          '100%': { opacity: '1', transform: 'translateX(0)' },
        },
        'slideOut': {
          '0%': { opacity: '1', transform: 'translateX(0)' },
          '100%': { opacity: '0', transform: 'translateX(10px)' },
        },
      },
    },
  },
  plugins: [],
}
`

// TailwindCSS provides base CSS including Datastar integration styles.
const TailwindCSS = `@tailwind base;
@tailwind components;
@tailwind utilities;

/* Datastar loading indicator styles */
/* These work with data-indicator:loading attribute */
[data-indicator] {
  transition: opacity 200ms ease-in;
}

/* General loading state styling */
.loading {
  cursor: wait;
}

.loading button,
.loading input[type="submit"] {
  pointer-events: none;
  opacity: 0.7;
}

/* Morph animation for patched elements */
.morph-in {
  animation: morphIn 0.3s ease-out;
}

.morph-out {
  animation: morphOut 0.3s ease-out;
}

@keyframes morphIn {
  0% { opacity: 0; transform: translateY(-10px); }
  100% { opacity: 1; transform: translateY(0); }
}

@keyframes morphOut {
  0% { opacity: 1; }
  100% { opacity: 0; }
}

/* Error styling */
.error {
  @apply bg-red-50 border border-red-200 text-red-700 px-4 py-3 rounded relative;
}

/* Success styling */
.success {
  @apply bg-green-50 border border-green-200 text-green-700 px-4 py-3 rounded relative;
}

/* Loading spinner */
.spinner {
  @apply animate-spin rounded-full h-5 w-5 border-2 border-gray-300 border-t-blue-600;
}

/* Mobile-first responsive utilities */
@layer utilities {
  .safe-top {
    padding-top: env(safe-area-inset-top);
  }
  .safe-bottom {
    padding-bottom: env(safe-area-inset-bottom);
  }
  .safe-left {
    padding-left: env(safe-area-inset-left);
  }
  .safe-right {
    padding-right: env(safe-area-inset-right);
  }
  .safe-area {
    padding-top: env(safe-area-inset-top);
    padding-bottom: env(safe-area-inset-bottom);
    padding-left: env(safe-area-inset-left);
    padding-right: env(safe-area-inset-right);
  }
}
`

// PackageJSON provides the default package.json for Tailwind setup.
const PackageJSON = `{
  "name": "irgo-app",
  "version": "1.0.0",
  "scripts": {
    "build:css": "tailwindcss -i ./assets/css/input.css -o ./assets/css/output.css --minify",
    "watch:css": "tailwindcss -i ./assets/css/input.css -o ./assets/css/output.css --watch"
  },
  "devDependencies": {
    "tailwindcss": "^3.4.0"
  }
}
`

// DatastarScript returns the script tag for Datastar.
// For mobile apps, this would be bundled locally.
const DatastarScript = `<script type="module" src="https://cdn.jsdelivr.net/gh/starfederation/datastar/bundles/datastar.js"></script>`

// BaseHTML provides a minimal HTML template with Datastar and Tailwind.
const BaseHTML = `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0, user-scalable=no, viewport-fit=cover">
    <meta name="mobile-web-app-capable" content="yes">
    <meta name="apple-mobile-web-app-capable" content="yes">
    <meta name="apple-mobile-web-app-status-bar-style" content="default">
    <title>{{.Title}}</title>
    <link rel="stylesheet" href="/assets/css/output.css">
    <script type="module" src="/assets/js/datastar.js"></script>
</head>
<body class="bg-gray-50 text-gray-900 safe-area">
    <div id="app">
        {{.Content}}
    </div>
</body>
</html>
`
