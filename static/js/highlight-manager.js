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
    const wordSelector = `[${DOM_ATTRIBUTES.word}="${escapeSelector(word)}"][${DOM_ATTRIBUTES.track}="${track}"]`;
    document.querySelectorAll(wordSelector).forEach(el => {
      this.highlightElement(el, categories, shouldAdd);
    });

    const wordCountElement = document.querySelector(`#wordCount${track}-${escapeSelector(word)}`);
    if (wordCountElement) {
      this.highlightElement(wordCountElement, categories, shouldAdd);
    }

    categories.forEach(category => {
      const categorySelector = `.${DOM_CLASSES.trackInfoItem}[${DOM_ATTRIBUTES.category}="${category}"][${DOM_ATTRIBUTES.trackIndex}="${track}"]`;
      const categoryElement = document.querySelector(categorySelector);
      if (categoryElement) {
        this.applyStyle(categoryElement, {
          backgroundColor: shouldAdd ? COLORS[category] : '#374151',
          color: shouldAdd ? '#000000' : '#d1d5db'
        });
      }
    });
  },

  toggleWordHighlight(element, shouldAdd = true) {
    const { word, track } = element.dataset;
    const categories = element.getAttribute(DOM_ATTRIBUTES.compromisePos)?.split(',') || [];
    this.highlightWordAndRelated(word, track, categories, shouldAdd);
  },

  togglePOSHighlight(category, trackIndex, shouldAdd) {
    const categorySelector = `.${DOM_CLASSES.trackInfoItem}[${DOM_ATTRIBUTES.category}="${category}"][${DOM_ATTRIBUTES.trackIndex}="${trackIndex}"]`;
    const categoryElement = document.querySelector(categorySelector);
    if (categoryElement) {
      this.applyStyle(categoryElement, {
        backgroundColor: shouldAdd ? COLORS[category] : '#374151',
        color: shouldAdd ? '#000000' : '#d1d5db'
      });
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
    document.querySelectorAll(`.${DOM_CLASSES.word}[${DOM_ATTRIBUTES.track}="${trackIndex}"]`)
      .forEach(word => {
        const categories = word.getAttribute(DOM_ATTRIBUTES.compromisePos)?.split(',') || [];
        if (categories.length > 0) {
          this.highlightElement(word, categories, shouldHighlight);
        }
      });
  }
};

window.highlightManager = highlightManager;