const trackInitializer = {
  initializeLyrics(lyricsElement, trackIndex) {
    if (!lyricsElement) return;

    const track = document.querySelector(`[id^="trackDetails${trackIndex}"]`);
    const ignoredWordsStr = track?.dataset.ignoredWords || '';
    const ignoredWords = ignoredWordsStr.split(',').map(w => w.trim().toLowerCase()).filter(Boolean);
    const ignoredWordsSet = new Set(ignoredWords);

    const words = lyricsElement.innerHTML
      .split(/\n/)
      .map(line => line.trim()
        .split(/(\s+)/)
        .map(word => {
          const cleanedWord = cleanWord(word);
          if (cleanedWord.length > 0 && ignoredWordsSet.has(cleanedWord)) {
            return word;
          }
          if (cleanedWord.length > 0) {
            return this.createWordSpan(word, cleanedWord, trackIndex);
          }
          return word;
        })
        .join('')
      )
      .join('\n');

    lyricsElement.innerHTML = words;
  },

  createWordSpan(originalWord, cleanedWord, trackIndex) {
    const wordCountElement = document.getElementById(`wordCount${trackIndex}-${cleanedWord}`);
    const count = wordCountElement?.getAttribute(DOM_ATTRIBUTES.count) || 0;

    return `<span class="${DOM_CLASSES.word}" 
      ${DOM_ATTRIBUTES.word}="${cleanedWord}" 
      ${DOM_ATTRIBUTES.track}="${trackIndex}" 
      ${DOM_ATTRIBUTES.count}="${count}" 
      onmouseover="highlightManager.toggleWordHighlight(this, true); tooltip.show(this)" 
      onmouseout="highlightManager.toggleWordHighlight(this, false); tooltip.hide()">${originalWord}</span>`;
  },

  attachWordCountListeners(trackElement) {
    trackElement.querySelectorAll('[id^="wordCount"]').forEach(wordCountElement => {
      const word = wordCountElement.getAttribute(DOM_ATTRIBUTES.word);
      const track = wordCountElement.getAttribute(DOM_ATTRIBUTES.track);
      const wordElement = trackElement.querySelector(
        `.${DOM_CLASSES.word}[${DOM_ATTRIBUTES.word}="${word}"][${DOM_ATTRIBUTES.track}="${track}"]`
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

  initializeTrack(trackIndex) {
    const lyricsElement = document.getElementById(`lyrics${trackIndex}`);
    if (!lyricsElement) return;

    this.initializeLyrics(lyricsElement, trackIndex);

    const trackElement = lyricsElement.closest('.track');
    if (!trackElement) return;

    const trackDetails = document.querySelector(`[id^="trackDetails${trackIndex}"]`);
    const ignoredWords = trackDetails?.dataset.ignoredWords || '';

    const posCategories = compromiseAnalyzer.analyze(lyricsElement.textContent, ignoredWords);
    const posContainer = compromiseAnalyzer.createPOSContainer(posCategories, trackIndex);

    const posContainerElement = trackElement.querySelector('.pos-container');
    if (posContainerElement) {
      posContainerElement.appendChild(posContainer);
    }

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

function getDebugInfo(trackIndex) {
  const lyricsElement = document.getElementById(`lyrics${trackIndex}`);
  if (!lyricsElement) return null;

  const trackDetails = document.querySelector(`[id^="trackDetails${trackIndex}"]`);
  const ignoredWords = trackDetails?.dataset.ignoredWords || '';
  const text = lyricsElement.textContent;

  const allTextWords = text.toLowerCase()
    .split(/[\s\n]+/)
    .map(w => w.trim())
    .filter(w => w.length > 0);

  const analysisResults = compromiseAnalyzer.analyze(text, ignoredWords);

  const processedWords = new Set();
  Object.values(analysisResults).forEach(category => {
    if (category.unique) {
      category.unique.forEach(word => processedWords.add(word.toLowerCase()));
    }
  });

  const missingWords = allTextWords.filter(word =>
    !processedWords.has(word) &&
    !ignoredWords.includes(word)
  );

  return JSON.stringify({
    trackIndex,
    analysis: analysisResults,
    text,
    ignoredWords,
    debug: {
      missingWords,
      allWords: allTextWords,
      processedWords: Array.from(processedWords)
    }
  }, null, 2);
}

function copyDebugInfo(trackIndex) {
  const debugInfo = getDebugInfo(trackIndex);
  if (!debugInfo) return;

  navigator.clipboard.writeText(debugInfo).then(() => {
    const button = document.querySelector(`#debugButton${trackIndex}`);
    const originalText = button.textContent;
    button.textContent = 'Copied!';
    setTimeout(() => {
      button.textContent = originalText;
    }, 2000);
  });
}
window.copyDebugInfo = copyDebugInfo;
window.trackInitializer = trackInitializer;