'use strict';

// ── State ────────────────────────────────────────────────────────────────────

let state = {
  recipes: [],
  activeCategory: null,
  searchQuery: '',
  editingId: null,
  pendingDeleteId: null,
  pendingRemoveImage: false,
  thumb: { x: 50, y: 50, zoom: 1 },
};

// Stored oven_mode values (English) mapped to their UI label keys.
const OVEN_MODE_LABELS = {
  fan: 'Fan oven',
  conventional: 'Conventional (top and bottom heat)',
};

// ── DOM refs ─────────────────────────────────────────────────────────────────

const $ = (sel, ctx = document) => ctx.querySelector(sel);
const $$ = (sel, ctx = document) => [...ctx.querySelectorAll(sel)];

const categoryList  = $('#category-list');
const recipeCards   = $('#recipe-cards');
const noRecipesMsg  = $('#no-recipes-msg');
const listHeading   = $('#list-heading');
const viewList      = $('#view-list');
const viewDetail    = $('#view-detail');
const recipeDetail  = $('#recipe-detail');
const searchInput   = $('#search-input');
const modalOverlay  = $('#modal-overlay');
const deleteOverlay = $('#delete-overlay');
const lightboxOverlay = $('#lightbox-overlay');
const lightboxImg   = $('#lightbox-img');
const recipeForm    = $('#recipe-form');

// ── Thumbnail crop ────────────────────────────────────────────────────────────

// Frame shape for list thumbnails (width / height). Keep in sync with the
// aspect-ratio of .recipe-card-thumb and .thumb-box in style.css.
const THUMB_ASPECT = 2;

// Lays an absolutely positioned <img> out inside a THUMB_ASPECT frame.
// x/y (0-100) pick which part of the overflow is shown (50 = centred, like
// object-position); zoom scales the cover-fitted image (1 = just covers the
// frame, below 1 = whole image visible, with the frame backdrop around it). Everything is a
// percentage of the frame, so it needs no resize handling. Defaults are a
// centred cover fit, so uncropped recipes look as they did before.
function applyThumbCrop(img, crop) {
  const x = crop?.x ?? 50;
  const y = crop?.y ?? 50;
  const zoom = crop?.zoom ?? 1;
  const apply = () => {
    const ratio = img.naturalWidth / img.naturalHeight;
    const wPct = zoom * Math.max(1, ratio / THUMB_ASPECT) * 100;
    const hPct = (wPct * THUMB_ASPECT) / ratio;
    img.style.width = `${wPct}%`;
    img.style.height = `${hPct}%`;
    img.style.left = `${-(wPct - 100) * x / 100}%`;
    img.style.top = `${-(hPct - 100) * y / 100}%`;
  };
  if (img.complete && img.naturalWidth) apply();
  else img.addEventListener('load', apply, { once: true });
}

// Blurred copy of the photo behind the thumbnail (see ::before in style.css),
// visible only where the zoomed-out image doesn't cover the frame.
function setThumbBackdrop(frame, src) {
  frame.style.setProperty('--thumb-src', `url(${JSON.stringify(src)})`);
}

// Lowest zoom that still shows the whole image inside the frame.
function minThumbZoom(img) {
  const ratio = img.naturalWidth / img.naturalHeight;
  return Math.max(0.05, Math.ceil(Math.min(ratio / THUMB_ASPECT, THUMB_ASPECT / ratio) * 100) / 100);
}

// ── Icon helper ───────────────────────────────────────────────────────────────

function icon(name) {
  const svg = document.createElementNS('http://www.w3.org/2000/svg', 'svg');
  svg.setAttribute('aria-hidden', 'true');
  svg.classList.add('md-icon');
  const use = document.createElementNS('http://www.w3.org/2000/svg', 'use');
  use.setAttribute('href', `#icon-${name}`);
  svg.appendChild(use);
  return svg;
}

// ── Category helpers ──────────────────────────────────────────────────────────

function getAllCategories(recipes) {
  const set = new Set();
  recipes.forEach(r => (r.categories || []).forEach(c => set.add(c.toLowerCase().trim())));
  return [...set].sort();
}

// ── Rendering ─────────────────────────────────────────────────────────────────

function renderCategories() {
  const categories = getAllCategories(state.recipes);

  categoryList.innerHTML = '';

  const allLi = document.createElement('li');
  const allA  = document.createElement('a');
  allA.href = '#';
  allA.textContent = t('All recipes');
  allA.className = state.activeCategory === null ? 'active' : '';
  allA.addEventListener('click', e => { e.preventDefault(); selectCategory(null); });
  allLi.appendChild(allA);
  categoryList.appendChild(allLi);

  categories.forEach(cat => {
    const li = document.createElement('li');
    const a  = document.createElement('a');
    a.href = '#';
    a.textContent = categoryLabel(cat);
    a.className = state.activeCategory === cat ? 'active' : '';
    a.addEventListener('click', e => { e.preventDefault(); selectCategory(cat); });
    li.appendChild(a);
    categoryList.appendChild(li);
  });
}

function getFilteredRecipes() {
  let recipes = state.recipes;

  if (state.activeCategory) {
    recipes = recipes.filter(r =>
      (r.categories || []).map(c => c.toLowerCase().trim()).includes(state.activeCategory));
  }

  if (state.searchQuery) {
    const q = state.searchQuery.toLowerCase();
    recipes = recipes.filter(r =>
      r.title.toLowerCase().includes(q) ||
      (r.description || '').toLowerCase().includes(q) ||
      (r.categories || []).some(c => c.toLowerCase().includes(q) || categoryLabel(c).toLowerCase().includes(q)) ||
      (r.ingredients || []).some(i => i.toLowerCase().includes(q))
    );
  }

  return recipes;
}

function renderRecipeList() {
  const recipes = getFilteredRecipes();
  recipeCards.innerHTML = '';

  if (recipes.length === 0) {
    noRecipesMsg.classList.remove('hidden');
  } else {
    noRecipesMsg.classList.add('hidden');
    recipes.forEach(r => recipeCards.appendChild(buildCard(r)));
  }

  if (state.searchQuery) {
    listHeading.textContent = `${t('Search')}: "${state.searchQuery}"`;
  } else if (state.activeCategory) {
    listHeading.textContent = categoryLabel(state.activeCategory);
  } else {
    listHeading.textContent = t('All recipes');
  }
}

function buildCard(recipe) {
  const card = document.createElement('div');
  card.className = 'recipe-card';
  card.setAttribute('data-id', recipe.id);

  const h3 = document.createElement('h3');
  h3.textContent = recipe.title;

  const p = document.createElement('p');
  p.textContent = recipe.description || '';

  const tags = document.createElement('div');
  tags.className = 'card-tags';
  (recipe.categories || []).forEach(c => {
    const span = document.createElement('span');
    span.className = 'tag';
    span.textContent = categoryLabel(c);
    tags.appendChild(span);
  });

  const actions = document.createElement('div');
  actions.className = 'card-actions admin-only';

  const editBtn = document.createElement('button');
  editBtn.className = 'btn-icon';
  editBtn.title = t('Edit recipe');
  editBtn.appendChild(icon('edit'));
  editBtn.addEventListener('click', e => { e.stopPropagation(); openEditModal(recipe.id); });

  const delBtn = document.createElement('button');
  delBtn.className = 'btn-icon';
  delBtn.title = t('Delete recipe');
  delBtn.appendChild(icon('delete'));
  delBtn.addEventListener('click', e => { e.stopPropagation(); openDeleteModal(recipe.id); });

  actions.appendChild(editBtn);
  actions.appendChild(delBtn);

  if (recipe.image_url) {
    const thumbFrame = document.createElement('div');
    thumbFrame.className = 'recipe-card-thumb';
    const thumb = document.createElement('img');
    thumb.alt = recipe.title;
    thumb.loading = 'lazy';
    applyThumbCrop(thumb, { x: recipe.thumb_x, y: recipe.thumb_y, zoom: recipe.thumb_zoom });
    setThumbBackdrop(thumbFrame, recipe.image_url);
    thumb.src = recipe.image_url;
    thumbFrame.appendChild(thumb);
    card.appendChild(thumbFrame);
  }

  card.appendChild(h3);
  card.appendChild(p);
  card.appendChild(tags);
  card.appendChild(actions);

  card.addEventListener('click', () => showDetail(recipe.id));

  return card;
}

function showDetail(id) {
  go(`#/recipes/${id}`);
}

function renderDetail(id) {
  const recipe = state.recipes.find(r => r.id === id);
  if (!recipe) return false;

  recipeDetail.innerHTML = '';

  const h2 = document.createElement('h2');
  h2.textContent = recipe.title;

  const meta = document.createElement('div');
  meta.className = 'detail-meta';
  (recipe.categories || []).forEach(c => {
    const span = document.createElement('span');
    span.className = 'tag';
    span.textContent = categoryLabel(c);
    meta.appendChild(span);
  });

  const detailActions = document.createElement('div');
  detailActions.className = 'detail-actions admin-only';
  const editBtn = document.createElement('button');
  editBtn.className = 'btn-icon';
  editBtn.title = t('Edit recipe');
  editBtn.appendChild(icon('edit'));
  editBtn.addEventListener('click', () => openEditModal(recipe.id));
  const delBtn = document.createElement('button');
  delBtn.className = 'btn-icon';
  delBtn.title = t('Delete recipe');
  delBtn.appendChild(icon('delete'));
  delBtn.addEventListener('click', () => openDeleteModal(recipe.id));
  detailActions.appendChild(editBtn);
  detailActions.appendChild(delBtn);

  const descSection = document.createElement('section');
  const descP = document.createElement('p');
  descP.textContent = recipe.description || '';
  descSection.appendChild(descP);

  const ingSection = document.createElement('section');
  const ingH3 = document.createElement('h3');
  ingH3.textContent = t('Ingredients');
  const ingUl = document.createElement('ul');
  (recipe.ingredients || []).forEach(ing => {
    const li = document.createElement('li');
    li.textContent = ing;
    ingUl.appendChild(li);
  });
  ingSection.appendChild(ingH3);
  ingSection.appendChild(ingUl);

  const stepsSection = document.createElement('section');
  const stepsH3 = document.createElement('h3');
  stepsH3.textContent = t('Steps');
  const stepsUl = document.createElement('ul');
  stepsUl.style.listStyle = 'none';
  stepsUl.style.padding = '0';
  (recipe.steps || []).forEach((step, i) => {
    const li = document.createElement('li');
    li.textContent = `${t('Step')} ${i + 1}: ${step}`;
    stepsUl.appendChild(li);
  });
  stepsSection.appendChild(stepsH3);
  stepsSection.appendChild(stepsUl);

  const titleRow = document.createElement('div');
  titleRow.className = 'detail-title-row';
  titleRow.appendChild(h2);
  titleRow.appendChild(detailActions);
  recipeDetail.appendChild(titleRow);
  recipeDetail.appendChild(meta);
  if (recipe.image_url) {
    const frame = document.createElement('button');
    frame.type = 'button';
    frame.className = 'recipe-photo-frame';
    frame.setAttribute('aria-label', t('View full image'));
    const bg = document.createElement('img');
    bg.src = recipe.image_url;
    bg.alt = '';
    bg.setAttribute('aria-hidden', 'true');
    bg.className = 'recipe-photo-bg';
    const img = document.createElement('img');
    img.src = recipe.image_url;
    img.alt = recipe.title;
    img.className = 'recipe-photo';
    frame.appendChild(bg);
    frame.appendChild(img);
    frame.addEventListener('click', () => openLightbox(recipe.image_url, recipe.title));
    recipeDetail.appendChild(frame);
  }
  recipeDetail.appendChild(descSection);
  recipeDetail.appendChild(ingSection);
  recipeDetail.appendChild(stepsSection);

  if (recipe.oven_temperature != null) {
    const ovenSection = document.createElement('section');
    const ovenH3 = document.createElement('h3');
    ovenH3.textContent = t('Oven Temperature');
    const ovenP = document.createElement('p');
    const ovenModeLabel = OVEN_MODE_LABELS[recipe.oven_mode];
    ovenP.textContent = `${recipe.oven_temperature}°C` + (ovenModeLabel ? ` · ${t(ovenModeLabel)}` : '');
    ovenSection.appendChild(ovenH3);
    ovenSection.appendChild(ovenP);
    recipeDetail.appendChild(ovenSection);
  }

  const links = recipe.links || [];
  if (links.length > 0) {
    const linksSection = document.createElement('section');
    const linksH3 = document.createElement('h3');
    linksH3.textContent = t('Related Links');
    const linksDiv = document.createElement('div');
    linksDiv.className = 'detail-links';

    links.forEach(link => {
      const a = document.createElement('a');
      if (link.type === 'recipe') {
        const linked = state.recipes.find(r => r.id === link.linked_recipe_id);
        a.appendChild(icon('link'));
        a.appendChild(document.createTextNode(' ' + (linked ? linked.title : t('Unknown recipe'))));
        a.href = '#';
        a.addEventListener('click', e => {
          e.preventDefault();
          if (linked) showDetail(linked.id);
        });
      } else {
        a.appendChild(icon('public'));
        a.appendChild(document.createTextNode(' ' + (link.label || link.url)));
        a.href = link.url;
        a.target = '_blank';
        a.rel = 'noopener noreferrer';
      }
      linksDiv.appendChild(a);
    });

    linksSection.appendChild(linksH3);
    linksSection.appendChild(linksDiv);
    recipeDetail.appendChild(linksSection);
  }

  showView('detail');
  return true;
}

// ── View switching ────────────────────────────────────────────────────────────

function showView(name) {
  const catHome       = $('#view-categories-home');
  const recipeBrowser = $('#recipe-browser');

  if (name === 'categories') {
    catHome.classList.remove('hidden');
    recipeBrowser.classList.add('hidden');
    return;
  }

  catHome.classList.add('hidden');
  recipeBrowser.classList.remove('hidden');

  viewList.classList.remove('active');
  viewList.classList.add('hidden');
  viewDetail.classList.remove('active');
  viewDetail.classList.add('hidden');

  if (name === 'list') {
    viewList.classList.add('active');
    viewList.classList.remove('hidden');
  } else {
    viewDetail.classList.add('active');
    viewDetail.classList.remove('hidden');
  }
}

function selectCategory(cat) {
  go(cat ? `#/categories/${encodeURIComponent(cat)}` : '#/recipes');
}

function renderList(cat) {
  state.activeCategory = cat;
  state.searchQuery = '';
  searchInput.value = '';
  renderCategories();
  renderRecipeList();
  showView('list');
  closeSidebar();
}

function showCategoriesHome() {
  go('#/');
}

function renderHome() {
  state.activeCategory = null;
  state.searchQuery = '';
  searchInput.value = '';
  renderCategoryCards();
  showView('categories');
}

function renderCategoryCards() {
  const categories = getAllCategories(state.recipes);
  const grid = $('#category-cards-grid');
  const msg  = $('#no-categories-msg');
  grid.innerHTML = '';

  if (categories.length === 0) {
    msg.classList.remove('hidden');
    return;
  }
  msg.classList.add('hidden');

  categories.forEach(cat => {
    const catRecipes = state.recipes.filter(r =>
      (r.categories || []).map(c => c.toLowerCase().trim()).includes(cat));

    const card = document.createElement('div');
    card.className = 'category-home-card';
    card.addEventListener('click', () => selectCategory(cat));

    const emoji = document.createElement('div');
    emoji.className = 'cat-emoji';
    emoji.textContent = getCategoryEmoji(cat);

    const h3 = document.createElement('h3');
    h3.textContent = categoryLabel(cat);

    const count = document.createElement('p');
    count.className = 'cat-count';
    count.textContent = `${catRecipes.length} ${t(catRecipes.length === 1 ? 'recipe' : 'recipes')}`;

    const preview = document.createElement('ul');
    preview.className = 'cat-preview';
    catRecipes.slice(0, 3).forEach(r => {
      const li = document.createElement('li');
      li.textContent = r.title;
      preview.appendChild(li);
    });

    card.appendChild(emoji);
    card.appendChild(h3);
    card.appendChild(count);
    card.appendChild(preview);
    grid.appendChild(card);
  });
}

function getCategoryEmoji(cat) {
  const map = {
    breakfast: '🍳', lunch: '🥙', dinner: '🍽', vegetarian: '🥦',
    vegan: '🌱', dessert: '🍰', snack: '🍿', soup: '🍲',
    pasta: '🍝', pizza: '🍕', salad: '🥗', meat: '🥩',
    fish: '🐟', seafood: '🦐', baking: '🥖', bread: '🍞',
    drinks: '🥤', cocktail: '🍹'
  };
  return map[cat.toLowerCase()] || '🍴';
}

// ── Modal: Add / Edit ─────────────────────────────────────────────────────────

function showThumbEditor(src, crop) {
  state.thumb = { x: crop?.x ?? 50, y: crop?.y ?? 50, zoom: crop?.zoom ?? 1 };
  const img = $('#form-image-preview-img');
  const zoom = $('#thumb-zoom');
  zoom.value = state.thumb.zoom;
  applyThumbCrop(img, state.thumb);
  setThumbBackdrop($('#thumb-box'), src);
  img.addEventListener('load', () => {
    zoom.min = Math.min(1, minThumbZoom(img));
    state.thumb.zoom = Math.max(state.thumb.zoom, parseFloat(zoom.min));
    zoom.value = state.thumb.zoom;
    applyThumbCrop(img, state.thumb);
  }, { once: true });
  img.src = src;
  $('#form-image-preview').classList.remove('hidden');
}

function hideThumbEditor() {
  $('#form-image-preview').classList.add('hidden');
}

function initThumbEditor() {
  const box = $('#thumb-box');
  const img = $('#form-image-preview-img');
  const refresh = () => applyThumbCrop(img, state.thumb);
  let drag = null;

  box.addEventListener('pointerdown', e => {
    drag = { px: e.clientX, py: e.clientY };
    box.setPointerCapture(e.pointerId);
    box.classList.add('dragging');
  });
  box.addEventListener('pointermove', e => {
    if (!drag) return;
    const b = box.getBoundingClientRect();
    const i = img.getBoundingClientRect();
    // Positive when the image is larger than the frame, negative when smaller
    // (then it moves within the free space instead).
    const overflowX = i.width - b.width;
    const overflowY = i.height - b.height;
    // Dragging the image right reveals more of its left side, so the focal % drops.
    if (Math.abs(overflowX) > 1) state.thumb.x = Math.min(100, Math.max(0, state.thumb.x - ((e.clientX - drag.px) / overflowX) * 100));
    if (Math.abs(overflowY) > 1) state.thumb.y = Math.min(100, Math.max(0, state.thumb.y - ((e.clientY - drag.py) / overflowY) * 100));
    drag.px = e.clientX;
    drag.py = e.clientY;
    refresh();
  });
  const end = () => { drag = null; box.classList.remove('dragging'); };
  box.addEventListener('pointerup', end);
  box.addEventListener('pointercancel', end);

  $('#thumb-zoom').addEventListener('input', e => {
    state.thumb.zoom = parseFloat(e.target.value);
    refresh();
  });

  let objectURL = null;
  $('#form-image').addEventListener('change', e => {
    if (objectURL) { URL.revokeObjectURL(objectURL); objectURL = null; }
    const file = e.target.files[0];
    if (!file) return;
    objectURL = URL.createObjectURL(file);
    state.pendingRemoveImage = false;
    showThumbEditor(objectURL, null);
  });
}

function openAddModal() {
  state.editingId = null;
  state.pendingRemoveImage = false;
  $('#modal-title').textContent = t('Add Recipe');
  recipeForm.reset();
  $('#form-id').value = '';
  clearDynamicList('ingredients-list');
  clearDynamicList('steps-list');
  clearDynamicList('links-list');
  addIngredientRow('');
  addStepRow('');
  hideThumbEditor();
  modalOverlay.classList.remove('hidden');
  $('#form-title').focus();
}

function openEditModal(id) {
  const recipe = state.recipes.find(r => r.id === id);
  if (!recipe) return;
  state.editingId = id;
  state.pendingRemoveImage = false;
  $('#modal-title').textContent = t('Edit Recipe');
  $('#form-id').value = id;
  $('#form-title').value = recipe.title;
  $('#form-description').value = recipe.description || '';
  $('#form-oven-temp').value = recipe.oven_temperature ?? '';
  $$('input[name="form-oven-mode"]').forEach(input => { input.checked = input.value === recipe.oven_mode; });
  $('#form-categories').value = (recipe.categories || []).join(', ');
  $('#form-tags').value = '';

  clearDynamicList('ingredients-list');
  (recipe.ingredients || []).forEach(ing => addIngredientRow(ing));
  if ((recipe.ingredients || []).length === 0) addIngredientRow('');

  clearDynamicList('steps-list');
  (recipe.steps || []).forEach(step => addStepRow(step));
  if ((recipe.steps || []).length === 0) addStepRow('');

  clearDynamicList('links-list');
  (recipe.links || []).forEach(link => addLinkRow(link));

  if (recipe.image_url) {
    showThumbEditor(recipe.image_url, { x: recipe.thumb_x, y: recipe.thumb_y, zoom: recipe.thumb_zoom });
  } else {
    hideThumbEditor();
  }
  $('#form-image').value = '';

  modalOverlay.classList.remove('hidden');
  $('#form-title').focus();
}

function openLightbox(src, alt) {
  lightboxImg.src = src;
  lightboxImg.alt = alt;
  lightboxOverlay.classList.remove('hidden');
}

function closeLightbox() {
  lightboxOverlay.classList.add('hidden');
  lightboxImg.removeAttribute('src');
}

function closeModal() {
  modalOverlay.classList.add('hidden');
  clearFormError();
}

function clearDynamicList(listId) {
  document.getElementById(listId).innerHTML = '';
}

function addIngredientRow(value) {
  const list = document.getElementById('ingredients-list');
  const row = buildTextRow(value, t('Ingredient'), () => row.remove());
  list.appendChild(row);
  if (!value) row.querySelector('input').focus();
}

function addStepRow(value) {
  const list = document.getElementById('steps-list');
  const row = buildTextRow(value, t('Step description'), () => row.remove());
  list.appendChild(row);
  if (!value) row.querySelector('input').focus();
}

function addLinkRow(link = {}) {
  const list = document.getElementById('links-list');
  const row = document.createElement('div');
  row.className = 'dynamic-item';

  const typeSelect = document.createElement('select');
  typeSelect.className = 'link-type';
  ['external', 'recipe'].forEach(t => {
    const opt = document.createElement('option');
    opt.value = t;
    opt.textContent = t === 'external' ? t('External URL') : t('Recipe');
    typeSelect.appendChild(opt);
  });
  typeSelect.value = link.type || 'external';

  const dlId = 'dl-' + Math.random().toString(36).slice(2);
  const datalist = document.createElement('datalist');
  datalist.id = dlId;

  const urlInput = document.createElement('input');
  urlInput.type = 'text';
  urlInput.className = 'link-url';

  if (link.type === 'recipe') {
    const linked = state.recipes.find(r => r.id === link.linked_recipe_id);
    urlInput.value = linked ? linked.title : '';
    urlInput.placeholder = t('Type recipe name…');
    urlInput.setAttribute('list', dlId);
    state.recipes.forEach(r => {
      const opt = document.createElement('option');
      opt.value = r.title;
      datalist.appendChild(opt);
    });
  } else {
    urlInput.value = link.url || '';
    urlInput.placeholder = 'URL';
  }

  const labelInput = document.createElement('input');
  labelInput.type = 'text';
  labelInput.className = 'link-label';
  labelInput.placeholder = t('Label (optional)');
  labelInput.value = link.label || '';

  typeSelect.addEventListener('change', () => {
    if (typeSelect.value === 'recipe') {
      urlInput.placeholder = t('Type recipe name…');
      urlInput.value = '';
      datalist.innerHTML = '';
      state.recipes.forEach(r => {
        const opt = document.createElement('option');
        opt.value = r.title;
        datalist.appendChild(opt);
      });
      urlInput.setAttribute('list', dlId);
      labelInput.style.display = 'none';
    } else {
      urlInput.placeholder = 'URL';
      urlInput.removeAttribute('list');
      labelInput.style.display = '';
    }
  });
  if (link.type === 'recipe') labelInput.style.display = 'none';

  const removeBtn = document.createElement('button');
  removeBtn.type = 'button';
  removeBtn.className = 'btn-icon';
  removeBtn.title = t('Remove');
  removeBtn.appendChild(icon('close'));
  removeBtn.addEventListener('click', () => row.remove());

  row.appendChild(typeSelect);
  row.appendChild(urlInput);
  row.appendChild(datalist);
  row.appendChild(labelInput);
  row.appendChild(removeBtn);
  list.appendChild(row);
}

let _dragSrc = null;
function makeDraggable(list) {
  list.addEventListener('dragstart', e => {
    _dragSrc = e.target.closest('.dynamic-item');
    e.dataTransfer.effectAllowed = 'move';
  });
  list.addEventListener('dragover', e => {
    e.preventDefault();
    const target = e.target.closest('.dynamic-item');
    if (target && target !== _dragSrc) target.classList.add('drag-over');
  });
  list.addEventListener('dragleave', e => {
    const target = e.target.closest('.dynamic-item');
    if (target) target.classList.remove('drag-over');
  });
  list.addEventListener('drop', e => {
    e.preventDefault();
    const target = e.target.closest('.dynamic-item');
    if (!target || target === _dragSrc) return;
    target.classList.remove('drag-over');
    const rect = target.getBoundingClientRect();
    if (e.clientY < rect.top + rect.height / 2) {
      list.insertBefore(_dragSrc, target);
    } else {
      list.insertBefore(_dragSrc, target.nextSibling);
    }
  });
  list.addEventListener('dragend', () => {
    list.querySelectorAll('.drag-over').forEach(el => el.classList.remove('drag-over'));
  });
}

function buildTextRow(value, placeholder, onRemove) {
  const row = document.createElement('div');
  row.className = 'dynamic-item';
  row.setAttribute('draggable', 'true');

  const handle = document.createElement('span');
  handle.className = 'drag-handle';
  handle.textContent = '⠿';
  handle.title = t('Drag to reorder');

  const input = document.createElement('input');
  input.type = 'text';
  input.value = value;
  input.placeholder = placeholder;

  const removeBtn = document.createElement('button');
  removeBtn.type = 'button';
  removeBtn.className = 'btn-icon';
  removeBtn.title = t('Remove');
  removeBtn.appendChild(icon('close'));
  removeBtn.addEventListener('click', onRemove);

  row.appendChild(handle);
  row.appendChild(input);
  row.appendChild(removeBtn);
  return row;
}

function collectDynamicValues(listId) {
  return $$(`#${listId} .dynamic-item input[type="text"]`)
    .map(i => i.value.trim())
    .filter(Boolean);
}

function collectLinks() {
  return $$('#links-list .dynamic-item').map(row => {
    const type = row.querySelector('.link-type').value;
    const urlVal = row.querySelector('.link-url').value.trim();
    const labelVal = row.querySelector('.link-label')?.value ?? '';
    if (!urlVal) return null;

    if (type === 'recipe') {
      const match = state.recipes.find(r =>
        r.title.toLowerCase() === urlVal.toLowerCase());
      return match ? { type: 'recipe', linked_recipe_id: match.id } : null;
    }
    return { type: 'external', url: urlVal, label: labelVal.trim() };
  }).filter(Boolean);
}

function showFormError(msg) {
  const el = $('#form-error');
  el.textContent = msg;
  el.classList.remove('hidden');
  el.scrollIntoView({ block: 'nearest' });
}

function clearFormError() {
  const el = $('#form-error');
  el.textContent = '';
  el.classList.add('hidden');
}

// Crop fields for the save request; null (default crop) when the form has no photo.
function thumbCropFields() {
  if ($('#form-image-preview').classList.contains('hidden')) {
    return { thumb_x: null, thumb_y: null, thumb_zoom: null };
  }
  const r = (n, d) => Math.round(n * d) / d;
  return { thumb_x: r(state.thumb.x, 10), thumb_y: r(state.thumb.y, 10), thumb_zoom: r(state.thumb.zoom, 100) };
}

async function handleFormSubmit(e) {
  e.preventDefault();
  clearFormError();

  const title = $('#form-title').value.trim();
  if (!title) { showFormError(t('Please enter a recipe title.')); $('#form-title').focus(); return; }

  const categories = $('#form-categories').value
    .split(',').map(c => c.trim().toLowerCase()).filter(Boolean);
  if (categories.length === 0) { showFormError(t('Please enter at least one category.')); $('#form-categories').focus(); return; }

  const ingredients = collectDynamicValues('ingredients-list');
  if (ingredients.length === 0) { showFormError(t('Please add at least one ingredient.')); return; }

  const steps = collectDynamicValues('steps-list');
  if (steps.length === 0) { showFormError(t('Please add at least one step.')); return; }

  const links = collectLinks();

  const ovenTempVal = $('#form-oven-temp').value;
  const ovenModeChecked = document.querySelector('input[name="form-oven-mode"]:checked');
  if (ovenTempVal && !ovenModeChecked) { showFormError(t('Please select an oven mode.')); return; }

  const recipeData = {
    title,
    description: $('#form-description').value.trim(),
    categories,
    ingredients,
    steps,
    links,
    oven_temperature: ovenTempVal ? parseInt(ovenTempVal, 10) : null,
    oven_mode: ovenTempVal ? ovenModeChecked.value : null,
    ...thumbCropFields(),
  };

  const editingId = state.editingId;

  try {
    let saved;
    if (editingId) {
      saved = await window.API.recipes.update(editingId, recipeData);
    } else {
      saved = await window.API.recipes.create(recipeData);
    }

    const imageFile = $('#form-image').files[0];
    if (imageFile) {
      await window.API.recipes.uploadImage(saved.id, imageFile);
    } else if (state.pendingRemoveImage) {
      await window.API.recipes.deleteImage(saved.id);
    }

    closeModal();
    await loadAndRender();
    if (!viewDetail.classList.contains('hidden') || !editingId) {
      const updated = state.recipes.find(r => r.id === saved.id);
      if (updated) showDetail(updated.id);
    }
  } catch (err) {
    showFormError(err.message || t('Failed to save recipe. Please try again.'));
  }
}

// ── Modal: Delete ─────────────────────────────────────────────────────────────

function openDeleteModal(id) {
  const recipe = state.recipes.find(r => r.id === id);
  if (!recipe) return;
  state.pendingDeleteId = id;
  $('#delete-recipe-name').textContent = recipe.title;

  const linkedBy = state.recipes.filter(r =>
    r.id !== id && (r.links || []).some(l => l.type === 'recipe' && l.linked_recipe_id === id)
  );
  const blockMsg = $('#delete-block-msg');
  const confirmBtn = $('#btn-confirm-delete');
  if (linkedBy.length > 0) {
    blockMsg.textContent = `${t('Cannot delete: linked by')} ${linkedBy.map(r => r.title).join(', ')}`;
    blockMsg.classList.remove('hidden');
    confirmBtn.disabled = true;
  } else {
    blockMsg.classList.add('hidden');
    confirmBtn.disabled = false;
  }

  deleteOverlay.classList.remove('hidden');
}

function closeDeleteModal() {
  state.pendingDeleteId = null;
  $('#btn-confirm-delete').disabled = false;
  $('#delete-block-msg').classList.add('hidden');
  deleteOverlay.classList.add('hidden');
}

async function confirmDelete() {
  if (!state.pendingDeleteId) return;

  const linkedBy = state.recipes.filter(r =>
    r.id !== state.pendingDeleteId &&
    (r.links || []).some(l => l.type === 'recipe' && l.linked_recipe_id === state.pendingDeleteId)
  );
  if (linkedBy.length > 0) return;

  try {
    await window.API.recipes.delete(state.pendingDeleteId);
    closeDeleteModal();
    await loadAndRender();
  } catch (err) {
    const blockMsg = $('#delete-block-msg');
    blockMsg.textContent = err.message || t('Failed to delete recipe.');
    blockMsg.classList.remove('hidden');
  }
}

// ── Utilities ─────────────────────────────────────────────────────────────────

function capitalize(str) {
  return str.charAt(0).toUpperCase() + str.slice(1);
}

// Categories are stored in English (e.g. "breakfast"); show the translated name.
function categoryLabel(cat) {
  return tCategory(cat) || capitalize(cat);
}

// ── Mobile sidebar drawer ─────────────────────────────────────────────────────

function toggleSidebar() {
  const sidebar = $('#sidebar');
  const overlay = $('#drawer-overlay');
  const isOpen = sidebar.classList.toggle('open');
  overlay.classList.toggle('hidden', !isOpen);
}

function closeSidebar() {
  $('#sidebar').classList.remove('open');
  $('#drawer-overlay').classList.add('hidden');
}

// ── Admin login ───────────────────────────────────────────────────────────────

const loginOverlay = $('#login-overlay');

function applyAuthState() {
  const admin = window.API.auth.isLoggedIn();
  document.body.classList.toggle('is-admin', admin);
  $('#btn-auth').textContent = t(admin ? 'Log out' : 'Admin login');
}

function openLoginModal() {
  $('#login-form').reset();
  $('#login-error').classList.add('hidden');
  loginOverlay.classList.remove('hidden');
  $('#login-pin').focus();
}

function closeLoginModal() {
  loginOverlay.classList.add('hidden');
}

async function handleLogin(e) {
  e.preventDefault();
  const errorEl = $('#login-error');
  errorEl.classList.add('hidden');
  try {
    await window.API.auth.login($('#login-pin').value.trim());
    closeLoginModal();
  } catch (err) {
    errorEl.textContent = err.status === 401 ? t('Invalid PIN') : (err.message || t('Login failed'));
    errorEl.classList.remove('hidden');
  }
}

// ── Event wiring ──────────────────────────────────────────────────────────────

function wireEvents() {
  $('#logo').addEventListener('click', showCategoriesHome);

  $('#btn-add-recipe').addEventListener('click', openAddModal);
  $('#btn-cancel').addEventListener('click', closeModal);
  recipeForm.addEventListener('submit', handleFormSubmit);
  recipeForm.addEventListener('keydown', e => {
    if (e.key === 'Enter' && e.target.tagName === 'INPUT') {
      e.preventDefault();
      e.target.blur();
    }
  });

  $('#btn-add-ingredient').addEventListener('click', () => addIngredientRow(''));
  $('#btn-add-step').addEventListener('click', () => addStepRow(''));
  $('#btn-add-link').addEventListener('click', () => addLinkRow());

  initThumbEditor();

  $('#btn-remove-image').addEventListener('click', () => {
    state.pendingRemoveImage = true;
    hideThumbEditor();
    $('#form-image').value = '';
  });

  $('#btn-convert-temp').addEventListener('click', () => {
    const f = parseFloat($('#form-oven-temp').value);
    if (!isNaN(f)) $('#form-oven-temp').value = Math.round((f - 32) * 5 / 9);
  });

  makeDraggable(document.getElementById('ingredients-list'));
  makeDraggable(document.getElementById('steps-list'));

  $('#btn-back').addEventListener('click', () => history.back());

  $('#btn-confirm-delete').addEventListener('click', confirmDelete);
  $('#btn-cancel-delete').addEventListener('click', closeDeleteModal);

  $('#btn-menu-toggle').addEventListener('click', toggleSidebar);
  $('#drawer-overlay').addEventListener('click', closeSidebar);

  $('#fab-add').addEventListener('click', openAddModal);

  $('#btn-auth').addEventListener('click', () => {
    if (window.API.auth.isLoggedIn()) window.API.auth.logout();
    else openLoginModal();
  });
  $('#login-form').addEventListener('submit', handleLogin);
  $('#btn-cancel-login').addEventListener('click', closeLoginModal);
  loginOverlay.addEventListener('click', e => { if (e.target === loginOverlay) closeLoginModal(); });
  window.addEventListener('auth-changed', applyAuthState);

  modalOverlay.addEventListener('click', e => { if (e.target === modalOverlay) closeModal(); });
  deleteOverlay.addEventListener('click', e => { if (e.target === deleteOverlay) closeDeleteModal(); });
  lightboxOverlay.addEventListener('click', closeLightbox);

  document.addEventListener('keydown', e => {
    if (e.key === 'Escape') { closeModal(); closeDeleteModal(); closeLoginModal(); closeLightbox(); closeSidebar(); }
  });

  searchInput.addEventListener('input', () => {
    state.searchQuery = searchInput.value.trim();
    if (state.searchQuery) {
      state.activeCategory = null;
      renderCategories();
      renderRecipeList();
      showView('list');
    } else {
      showCategoriesHome();
    }
  });
}

// ── Navigation (hash routes) ──────────────────────────────────────────────────
//   #/                      categories home
//   #/recipes               all recipes
//   #/categories/<name>     recipes in a category
//   #/recipes/<id>          recipe detail

function go(hash) {
  if (location.hash === hash) route();
  else location.hash = hash;
}

function route() {
  closeSidebar();
  state.searchQuery = '';
  searchInput.value = '';

  const parts = location.hash.replace(/^#\/?/, '').split('/').filter(Boolean);

  if (parts[0] === 'recipes' && parts.length === 1) {
    renderList(null);
  } else if (parts[0] === 'recipes') {
    if (!renderDetail(parseInt(parts[1], 10))) go('#/recipes');
  } else if (parts[0] === 'categories' && parts[1]) {
    let cat = parts[1];
    try { cat = decodeURIComponent(cat); } catch { /* keep raw */ }
    renderList(cat.toLowerCase().trim());
  } else {
    renderHome();
  }
}

window.addEventListener('hashchange', route);

// ── Data loading ──────────────────────────────────────────────────────────────

async function loadAndRender() {
  try {
    state.recipes = await window.API.recipes.getAll() || [];
  } catch {
    state.recipes = [];
  }
  renderCategories();
  renderCategoryCards();
  renderRecipeList();
  route();
}

// ── Bootstrap ─────────────────────────────────────────────────────────────────

async function init() {
  translateDOM();
  wireEvents();
  applyAuthState();
  await loadAndRender();
}

document.addEventListener('DOMContentLoaded', () => init());
