document.addEventListener('DOMContentLoaded', () => {
  const label = document.querySelector('.hero-label .mono')
  if (!label) return

  const text = label.textContent.trim()
  label.textContent = ''

  let i = 0
  const type = () => {
    if (i < text.length) {
      label.textContent += text[i++]
      setTimeout(type, 40)
    }
  }

  setTimeout(type, 300)
})