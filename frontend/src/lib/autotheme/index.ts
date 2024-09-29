function getPreferredTheme() {
  return window.matchMedia('(prefers-color-scheme: dark)').matches ? 'dark' : 'light'
}

function setTheme(theme: string) {
  document.documentElement.setAttribute('data-bs-theme', theme)
}

export default function() {
  setTheme(getPreferredTheme())

  window.matchMedia('(prefers-color-scheme: dark)').addEventListener('change', () => {
    setTheme(getPreferredTheme())
  })
}
