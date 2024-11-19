const trackInitializer = {
  initializeLyrics(lyricsElement, trackIndex, posCategorization) {
    if (!lyricsElement) return;

    const words = lyricsElement.innerHTML
      .split(/\n/)
      .map(line =>
        line.trim()
          .split(/(\s+)/)
          .map(word => {
            const cleanedWord = cleanWord(word);
            if (cleanedWord.length > 0) {
              return this.createWordSpan(cleanedWord, trackIndex, posCategorization);
            }
            return word;
          })
          .join('')
      )
      .join('\n');

    lyricsElement.innerHTML = words;
  },

  createWordSpan(cleanedWord, trackIndex, posCategorization) {
    const posCategory = posCategorization[cleanedWord] || '';
    const wordCountElement = document.getElementById(`wordCount${trackIndex}-${escapeSelector(cleanedWord)}`);
    const count = wordCountElement?.getAttribute(DOM_ATTRIBUTES.count) || 0;

    return `<span class="${DOM_CLASSES.word}" 
                  ${DOM_ATTRIBUTES.word}="${cleanedWord}" 
                  ${DOM_ATTRIBUTES.track}="${trackIndex}" 
                  ${DOM_ATTRIBUTES.count}="${count}" 
                  data-pos-category="${posCategory}" 
                  onmouseover="highlightManager.toggleWordHighlight(this, true); tooltip.show(this)" 
                  onmouseout="highlightManager.toggleWordHighlight(this, false); tooltip.hide()">${cleanedWord}</span>`;
  },

  attachWordCountListeners(trackElement) {
    trackElement.querySelectorAll('[id^="wordCount"]').forEach(wordCountElement => {
      const word = wordCountElement.getAttribute(DOM_ATTRIBUTES.word);
      const track = wordCountElement.getAttribute(DOM_ATTRIBUTES.track);
      const wordElement = trackElement.querySelector(
        `.${DOM_CLASSES.word}[${DOM_ATTRIBUTES.word}="${escapeSelector(word)}"][${DOM_ATTRIBUTES.track}="${track}"]`
      );

      if (wordElement) {
        const categories = wordElement.getAttribute(DOM_ATTRIBUTES.compromisePos)?.split(',') || [];

        wordCountElement.addEventListener('mouseover', () => {
          highlightManager.highlightWordAndRelated(word, track, categories, true);
        });

        wordCountElement.addEventListener('mouseout', () => {
          highlightManager.highlightWordAndRelated(word, track, categories, false);
        });
      }
    });
  },

  initializeTrack(trackIndex, posCategorization) {
    const lyricsElement = document.getElementById(`lyrics${trackIndex}`);
    if (!lyricsElement) return;

    this.initializeLyrics(lyricsElement, trackIndex, posCategorization);

    const trackElement = lyricsElement.closest('.track');
    if (!trackElement) return;

    const posCategories = compromiseAnalyzer.analyze(lyricsElement.textContent);
    const posContainer = compromiseAnalyzer.createPOSContainer(posCategories, trackIndex);
    trackElement.querySelector('.track-info-card')?.appendChild(posContainer);

    compromiseAnalyzer.attachToTrack(trackElement, trackIndex, posCategories);
    this.attachWordCountListeners(trackElement);
  },

  setupHighlightAllCheckboxes() {
    document.querySelectorAll('.highlight-all-checkbox').forEach(checkbox => {
      checkbox.addEventListener('change', function () {
        const trackIndex = this.getAttribute(DOM_ATTRIBUTES.track);
        highlightManager.highlightAllWords(trackIndex, this.checked);
      });
    });
  },

  initialize(trackData) {
    tooltip.init();
    trackData.forEach(track => {
      this.initializeTrack(track.index, track.posCategorization);
    });
    this.setupHighlightAllCheckboxes();
  }
};

window.trackInitializer = trackInitializer;