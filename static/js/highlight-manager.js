const highlightManager = {
  applyStyle(element, { backgroundColor = '', backgroundImage = '', color = '' }) {
    element.style.backgroundColor = backgroundColor;
    element.style.backgroundImage = backgroundImage;
    element.style.color = color;
  },

  createGradient(categories) {
    return `linear-gradient(45deg, ${categories.map((cat, i) => {
      const percentage = (i * 100) / categories.length;
      const nextPercentage = ((i + 1) * 100) / categories.length;
      return `${COLORS[cat]} ${percentage}%, ${COLORS[cat]} ${nextPercentage}%`;
    }).join(', ')
      })`;
  },

  highlightElement(element, categories, shouldAdd) {
    if (!shouldAdd) {
      this.applyStyle(element, {});
      return;
    }

    if (categories.length === 1) {
      this.applyStyle(element, {
        backgroundColor: COLORS[categories[0]],
        color: '#000000'
      });
    } else if (categories.length > 1) {
      this.applyStyle(element, {
        backgroundImage: this.createGradient(categories),
        color: '#000000'
      });
    }
  },

  highlightWordAndRelated(word, track, categories, shouldAdd = true) {
    const highlightAllCheckbox = document.querySelector(`input[id="highlightAll${track}"]`);
    if (highlightAllCheckbox?.checked) {
      shouldAdd = true;
    }

    const wordSelector = `[${DOM_ATTRIBUTES.word}="${word}"][${DOM_ATTRIBUTES.track}="${track}"]`;
    document.querySelectorAll(wordSelector).forEach(el => {
      this.highlightElement(el, categories, shouldAdd);
    });

    const wordCountElement = document.querySelector(`#wordCount${track}-${escapeSelector(word)}`);
    if (wordCountElement) {
      this.highlightElement(wordCountElement, categories, shouldAdd);
    }

    categories.forEach(category => {
      const posTagSelector = `.pos-tag[${DOM_ATTRIBUTES.category}="${category}"][${DOM_ATTRIBUTES.trackIndex}="${track}"]`;
      const posTagElement = document.querySelector(posTagSelector);
      if (posTagElement) {
        const baseColor = COLORS[category];
        if (shouldAdd) {
          posTagElement.style.transform = 'scale(1.09)';
          posTagElement.style.boxShadow = '10px 20px 10px rgba(0, 0, 0, 0.6)';
          posTagElement.style.background = `linear-gradient(135deg, ${baseColor}60, ${baseColor}70)`;
        } else {
          posTagElement.style.transform = '';
          posTagElement.style.boxShadow = '';
          posTagElement.style.background = `linear-gradient(135deg, ${baseColor}40, ${baseColor}50)`;
        }
      }
    });
  },

  toggleWordHighlight(element, shouldAdd = true) {
    const { word, track } = element.dataset;
    const categories = element.getAttribute(DOM_ATTRIBUTES.compromisePos)?.split(',') || [];
    this.highlightWordAndRelated(word, track, categories, shouldAdd);
  },

  togglePOSHighlight(category, trackIndex, shouldAdd) {
    const posTagSelector = `.pos-tag[${DOM_ATTRIBUTES.category}="${category}"][${DOM_ATTRIBUTES.trackIndex}="${trackIndex}"]`;
    const posTagElement = document.querySelector(posTagSelector);
    if (posTagElement) {
      const baseColor = COLORS[category];
      if (shouldAdd) {
        posTagElement.style.transform = 'scale(1.05)';
        posTagElement.style.boxShadow = '0 4px 6px rgba(0, 0, 0, 0.1)';
        posTagElement.style.background = `linear-gradient(135deg, ${baseColor}60, ${baseColor}70)`;
      } else {
        posTagElement.style.transform = '';
        posTagElement.style.boxShadow = '';
        posTagElement.style.background = `linear-gradient(135deg, ${baseColor}40, ${baseColor}50)`;
      }
    }

    const wordSelector = `.${DOM_CLASSES.word}[${DOM_ATTRIBUTES.track}="${trackIndex}"]`;
    document.querySelectorAll(wordSelector).forEach(element => {
      const categories = element.getAttribute(DOM_ATTRIBUTES.compromisePos)?.split(',') || [];
      if (categories.includes(category)) {
        const word = element.getAttribute(DOM_ATTRIBUTES.word);
        this.highlightElement(element, [category], shouldAdd);

        const wordCountElement = document.querySelector(`#wordCount${trackIndex}-${escapeSelector(word)}`);
        if (wordCountElement) {
          this.highlightElement(wordCountElement, [category], shouldAdd);
        }
      }
    });
  },

  highlightAllWords(trackIndex, shouldHighlight) {
    const wordElements = document.querySelectorAll(`.${DOM_CLASSES.word}[${DOM_ATTRIBUTES.track}="${trackIndex}"]`);

    wordElements.forEach(word => {
      const categories = word.getAttribute(DOM_ATTRIBUTES.compromisePos)?.split(',') || [];
      if (categories.length > 0) {
        const wordText = word.getAttribute(DOM_ATTRIBUTES.word);
        this.highlightWordAndRelated(wordText, trackIndex, categories, shouldHighlight);
      }
    });
  }
};

window.highlightManager = highlightManager;