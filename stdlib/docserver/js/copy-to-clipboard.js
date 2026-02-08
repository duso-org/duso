// Add copy-to-clipboard functionality to code blocks
document.addEventListener('DOMContentLoaded', function() {
  const codeBlocks = document.querySelectorAll('pre > code');

  codeBlocks.forEach((codeBlock) => {
    // Create the copy button container
    const copyButton = document.createElement('button');
    copyButton.className = 'copy-btn';
    copyButton.setAttribute('aria-label', 'Copy code to clipboard');
    copyButton.innerHTML = `
      <svg viewBox="0 0 24 24" width="18" height="18" stroke="currentColor" fill="none" stroke-linecap="round" stroke-linejoin="round">
        <rect x="9" y="9" width="13" height="13" rx="2" ry="2"></rect>
        <path d="M5 15H4a2 2 0 0 1-2-2V4a2 2 0 0 1 2-2h9a2 2 0 0 1 2 2v1"></path>
      </svg>
    `;

    copyButton.addEventListener('click', function() {
      const text = codeBlock.textContent;
      navigator.clipboard.writeText(text).then(() => {
        // Create and show feedback tooltip
        const tooltip = document.createElement('div');
        tooltip.className = 'copy-tooltip';
        tooltip.textContent = 'Copied';
        copyButton.parentElement.appendChild(tooltip);

        // Fade out and remove
        setTimeout(() => {
          tooltip.classList.add('fade-out');
          setTimeout(() => {
            tooltip.remove();
          }, 300);
        }, 1700);
      });
    });

    // Wrap the code block and add the button
    const preBlock = codeBlock.parentElement;
    const wrapper = document.createElement('div');
    wrapper.className = 'code-block-wrapper';
    preBlock.parentNode.insertBefore(wrapper, preBlock);
    wrapper.appendChild(preBlock);
    wrapper.appendChild(copyButton);
  });
});
