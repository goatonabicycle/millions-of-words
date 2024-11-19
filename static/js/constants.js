const COLORS = {
  noun: '#ffb3aa',
  verb: '#90caf9',
  adjective: '#a5d6a7',
  adverb: '#fff59d',
  pronoun: '#e1bee7',
  preposition: '#ffcc80',
  conjunction: '#ff8a65',
  determiner: '#80deea',
  auxiliary: '#ce93d8',
  interjection: '#ffab91'
};

const DOM_CLASSES = {
  word: 'word',
  trackInfoItem: 'track-info-item',
  highlightedAll: 'highlighted-all'
};

const DOM_ATTRIBUTES = {
  word: 'data-word',
  track: 'data-track',
  category: 'data-category',
  trackIndex: 'data-track-index',
  count: 'data-count',
  compromisePos: 'data-compromise-pos'
};

const cleanWord = (word) => word.toLowerCase().replace(/^[^a-zA-Z0-9À-ž'-]+|[^a-zA-Z0-9À-ž'-]+$/g, '');
const escapeSelector = (word) => word.replace(/([!"#$%&'()*+,.\/:;<=>?@[\\\]^`{|}~])/g, '\\$1');