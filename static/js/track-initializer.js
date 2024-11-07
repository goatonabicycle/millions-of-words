const trackInitializer = {
  initializeLyrics(lyricsElement, trackIndex, posCategorization) {
    if (!lyricsElement) return;

    const words = lyricsElement.innerHTML.split(/\n/).map(line =>
      line.trim().split(/(\s+)/).map(word => {
        const cleanedWord = cleanWord(word);
        if (cleanedWord.length > 0) {
          return this.createWordSpan(cleanedWord, trackIndex, posCategorization);
        }
        return word;
      }).join('')
    ).join('\n');

    lyricsElement.innerHTML = words;
  },

  createWordSpan(cleanedWord, trackIndex, posCategorization) {
    const posCategory = posCategorization[cleanedWord] || '';
    const wordCountElement = document.getElementById(`wordCount${trackIndex}-${escapeSelector(cleanedWord)}`);
    const count = wordCountElement?.getAttribute('data-count') || 0;

    return `<span class="word" 
          data-word="${cleanedWord}" 
          data-track="${trackIndex}" 
          data-count="${count}" 
          data-pos-category="${posCategory}" 
          onmouseover="highlightManager.toggleWordHighlight(this, true); tooltip.show(this)" 
          onmouseout="highlightManager.toggleWordHighlight(this, false); tooltip.hide()">${cleanedWord}</span>`;
  },

  attachWordCountListeners(trackElement) {
    trackElement.querySelectorAll('[id^="wordCount"]').forEach(wordCountElement => {
      wordCountElement.addEventListener('mouseover', () =>
        highlightManager.toggleWordHighlight(wordCountElement, true));
      wordCountElement.addEventListener('mouseout', () =>
        highlightManager.toggleWordHighlight(wordCountElement, false));
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
        const trackIndex = this.getAttribute('data-track');
        document.querySelectorAll(`.word[data-track="${trackIndex}"]`)
          .forEach(word => word.classList.toggle('highlighted-all', this.checked));
      });
    });
  },

  initialize(trackData) {
    tooltip.init();

    // Initialize each track
    trackData.forEach(track => {
      this.initializeTrack(track.index, track.posCategorization);
    });

    this.setupHighlightAllCheckboxes();
  }
};

window.trackInitializer = trackInitializer;