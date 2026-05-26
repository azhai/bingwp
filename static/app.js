// Bing Wallpaper - Mithril.js SPA
(function () {
  'use strict';

  var MIN_YEAR = 2009;
  var MIN_MONTH = 7;

  // ==================== State ====================
  var state = {
    year: new Date().getFullYear(),
    month: new Date().getMonth() + 1,
    wallpapers: [],
    monthLabel: '',
    prevMonth: '',
    nextMonth: '',
    loading: false,
    showPicker: false,
    pickerYear: new Date().getFullYear(),
    currentRoute: '',
  };

  // ==================== API ====================
  function fetchWallpapers(year, month) {
    state.loading = true;
    state.year = year;
    state.month = month;
    state.monthLabel = year + '\u5E74' + String(month).padStart(2, '0') + '\u6708';
    state.prevMonth = '';
    state.nextMonth = '';
    m.request({
      method: 'GET',
      url: '/api/wallpapers?year=' + year + '&month=' + month,
    }).then(function (data) {
      state.wallpapers = data.wallpapers || [];
      state.monthLabel = data.monthLabel;
      state.prevMonth = data.prevMonth;
      state.nextMonth = data.nextMonth;
      state.loading = false;
      state.showPicker = false;
      m.redraw();
    }).catch(function () {
      state.loading = false;
      m.redraw();
    });
  }

  function parseYM(ym) {
    if (!ym || ym === '/') return null;
    ym = ym.replace(/^\//, '');
    if (ym.length === 6) {
      var y = parseInt(ym.substring(0, 4));
      var m = parseInt(ym.substring(4, 6));
      if (m >= 1 && m <= 12) return { year: y, month: m };
    }
    return null;
  }

  function isMonthDisabled(year, month) {
    var now = new Date();
    if (year < MIN_YEAR) return true;
    if (year === MIN_YEAR && month < MIN_MONTH) return true;
    if (year > now.getFullYear()) return true;
    if (year === now.getFullYear() && month > now.getMonth() + 1) return true;
    return false;
  }

  function loadFromRoute(attrs) {
    var ym = null;
    if (attrs && attrs.ym) {
      ym = parseYM(attrs.ym);
    }
    var routeKey = ym ? (String(ym.year) + String(ym.month).padStart(2, '0')) : 'current';
    if (routeKey === state.currentRoute && state.wallpapers.length > 0) return;
    state.currentRoute = routeKey;
    if (ym) {
      fetchWallpapers(ym.year, ym.month);
    } else {
      fetchWallpapers(state.year, state.month);
    }
  }

  // ==================== Components ====================

  // Header
  var Header = {
    view: function () {
      return m('header.Header', [
        m('.container', [
          m('.Header-primary', [
            m('h1.Header-title', [
              m(m.route.Link, { href: '/' + String(new Date().getFullYear()) + String(new Date().getMonth() + 1).padStart(2, '0') }, 'Bing Wallpaper'),
            ]),
          ]),
          m('.Header-secondary', [
            m('.MonthNav', [
              state.prevMonth
                ? m(m.route.Link, {
                    href: state.prevMonth,
                    class: 'MonthNav-link MonthNav-prev',
                  }, m('span.icon', '\u2039'), ' \u4E0A\u6708')
                : m('span.MonthNav-link.MonthNav-disabled', '\u2039 \u4E0A\u6708'),
              m('button.MonthNav-picker', {
                onclick: function (e) {
                  e.stopPropagation();
                  state.showPicker = !state.showPicker;
                  if (state.showPicker) state.pickerYear = state.year;
                },
              }, state.monthLabel + ' \u25BE'),
              state.nextMonth
                ? m(m.route.Link, {
                    href: state.nextMonth,
                    class: 'MonthNav-link MonthNav-next',
                  }, '\u4E0B\u6708 ', m('span.icon', '\u203A'))
                : m('span.MonthNav-link.MonthNav-disabled', '\u4E0B\u6708 \u203A'),
            ]),
          ]),
        ]),
      ]);
    },
  };

  // Month Picker (dropdown)
  var MonthPicker = {
    view: function () {
      var now = new Date();
      var years = [];
      for (var y = now.getFullYear(); y >= MIN_YEAR; y--) years.push(y);

      return m('.MonthPicker.container', { onclick: function (e) { e.stopPropagation(); } }, [
        m('.MonthPicker-section', [
          m('.MonthPicker-label', '\u5E74\u4EFD'),
          m('.MonthPicker-years', years.map(function (year) {
            return m('button.MonthPicker-year', {
              class: year === state.pickerYear ? 'active' : '',
              onclick: function () { state.pickerYear = year; },
            }, year);
          })),
        ]),
        m('.MonthPicker-section', [
          m('.MonthPicker-label', '\u6708\u4EFD'),
          m('.MonthPicker-months', [12,11,10,9,8,7,6,5,4,3,2,1].map(function (month) {
            var disabled = isMonthDisabled(state.pickerYear, month);
            var isCurrent = state.pickerYear === state.year && month === state.month;
            if (disabled) {
              return m('span.MonthPicker-month.disabled', month + '\u6708');
            }
            var ym = String(state.pickerYear) + String(month).padStart(2, '0');
            return m('button.MonthPicker-month' + (isCurrent ? ' active' : ''), {
              onclick: function () { state.showPicker = false; m.route.set('/' + ym); },
            }, month + '\u6708');
          })),
        ]),
      ]);
    },
  };

  // Wallpaper Card
  var WallpaperCard = {
    view: function (vnode) {
      var wp = vnode.attrs.wp;
      return m('.WallpaperCard', [
        m('.WallpaperCard-image', [
          m('a', { href: wp.bingUrl, target: '_blank', rel: 'noopener' }, [
            m('img', {
              src: '/thumbs/' + wp.thumbnailPath,
              alt: wp.title,
              loading: 'lazy',
            }),
          ]),
        ]),
        m('.WallpaperCard-info', [
          m('.WallpaperCard-meta', [
            wp.headline ? m('span.WallpaperCard-headline', wp.headline) : null,
            m('span.WallpaperCard-date', wp.date),
          ]),
          m('h3.WallpaperCard-title', wp.title),
          wp.description ? m('.WallpaperCard-tooltip', [
            m('.WallpaperCard-tooltip-title', wp.title),
            m('.WallpaperCard-tooltip-date', wp.date),
            m('.WallpaperCard-tooltip-desc', wp.description),
          ]) : null,
        ]),
      ]);
    },
  };

  // Index Page
  var IndexPage = {
    oninit: function (vnode) {
      loadFromRoute(vnode.attrs);
    },
    onbeforeupdate: function (vnode) {
      var ym = parseYM(vnode.attrs.ym);
      var routeKey = ym ? (String(ym.year) + String(ym.month).padStart(2, '0')) : 'current';
      if (routeKey !== state.currentRoute) {
        loadFromRoute(vnode.attrs);
      }
      return true;
    },
    view: function () {
      return m('.IndexPage', [
        m(Header),
        state.showPicker ? m(MonthPicker) : null,
        m('.container.IndexPage-content', [
          state.loading
            ? m('.LoadingIndicator', m('.LoadingIndicator-spinner'))
            : state.wallpapers.length === 0
              ? m('.EmptyState', [
                  m('p', '\u8BE5\u6708\u4EFD\u6682\u65E0\u58C1\u7EB8\u6570\u636E'),
                  m('p', '\u8BF7\u8FD0\u884C ', m('code', './bingwp update'), ' \u66F4\u65B0\u6570\u636E'),
                ])
              : m('.WallpaperList', state.wallpapers.map(function (wp) {
                  return m(WallpaperCard, { wp: wp });
                })),
        ]),
      ]);
    },
  };

  // ==================== Router ====================
  document.addEventListener('click', function () {
    if (state.showPicker) {
      state.showPicker = false;
      m.redraw();
    }
  });

  m.route(document.getElementById('app'), '/' + String(state.year) + String(state.month).padStart(2, '0'), {
    '/:ym': IndexPage,
  });
})();
