const trackInitializer = {
  initializeLyrics(lyricsElement, trackIndex) {
    if (!lyricsElement) return;

    const track = document.querySelector(`[id^="trackDetails${trackIndex}"]`);

    const ignoredWords = (track?.dataset.ignoredWords || '').split(',').map(w => w.trim()).filter(Boolean);
    const patterns = new Set(ignoredWords.filter(w => /[()[\]{}:]/.test(w)));
    const exactWords = new Set(ignoredWords.filter(w => !/[()[\]{}:]/.test(w)));

    const expandedIgnoredSet = new Set(exactWords);
    exactWords.forEach(word => {
      const cleaned = cleanWord(word);
      if (cleaned) expandedIgnoredSet.add(cleaned);
    });

    const lines = lyricsElement.innerHTML.split(/\n/);
    const processedLines = lines.map(line => {
      let result = line;

      patterns.forEach(ignored => {
        const escapedIgnored = ignored.replace(/[.*+?^${}()|[\]\\]/g, '\\$&');
        const regex = new RegExp(`(${escapedIgnored})`, 'g');
        result = result.replace(regex, (match) => match);
      });

      const words = result.split(/(\s+)/);
      return words.map(segment => {
        if (!segment.trim()) return segment;

        if (patterns.has(segment)) return segment;

        const cleanedWord = cleanWord(segment);
        if (!cleanedWord) return segment;

        if (expandedIgnoredSet.has(cleanedWord)) return segment;

        return this.createWordSpan(segment, cleanedWord, trackIndex);
      }).join('');
    });

    lyricsElement.innerHTML = processedLines.join('\n');

    const wordCountContainer = track.querySelector('.word-counts');
    if (wordCountContainer) {
      const wordCountElements = wordCountContainer.querySelectorAll('[id^="wordCount"]');
      wordCountElements.forEach(element => {
        const word = element.getAttribute('data-word');
        if (expandedIgnoredSet.has(word)) {
          element.remove();
        }
      });
    }
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

function copyIgnoredWordsInfo(trackIndex) {
  const trackElement = document.querySelector(`[id^="trackDetails${trackIndex}"]`);
  const ignoredWordsStr = trackElement?.dataset.ignoredWords || '';
  const lyrics = document.getElementById(`lyrics${trackIndex}`).innerHTML;

  const info = {
    trackIndex,
    ignoredWords: ignoredWordsStr,
    ignoredWordsList: ignoredWordsStr.split(',').map(w => w.trim()).filter(Boolean),
    trackHTML: lyrics,
    wordElements: Array.from(document.querySelectorAll(`.word[data-track="${trackIndex}"]`))
      .map(el => ({
        word: el.getAttribute('data-word'),
        isInteractive: true
      })),
    textNodes: Array.from(document.getElementById(`lyrics${trackIndex}`).childNodes)
      .filter(node => node.nodeType === 3)
      .map(node => node.textContent.trim())
      .filter(text => text.length > 0)
  };

  navigator.clipboard.writeText(JSON.stringify(info, null, 2)).then(() => {
    const button = document.querySelector(`button[onclick="copyIgnoredWordsInfo(${trackIndex})"]`);
    button.textContent = 'Copied!';
    setTimeout(() => {
      button.textContent = 'Debug Ignored Words';
    }, 2000);
  });
}
window.copyIgnoredWordsInfo = copyIgnoredWordsInfo;