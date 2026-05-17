// Appearance preferences (palette + theme) live entirely on the frontend —
// they're UI presentation choices the backend doesn't need to know about.
// Both are persisted in localStorage and reflected as data-palette /
// data-theme attributes on <html>, which is where tokens.css reads them.

const PALETTE_KEY = 'pricehike.palette';
const THEME_KEY = 'pricehike.theme';

export const PALETTES = ['indigo', 'violet', 'teal', 'emerald', 'amber', 'crimson'] as const;
export type Palette = (typeof PALETTES)[number];

export const PALETTE_LABELS: Record<Palette, string> = {
	indigo: 'Indigo',
	violet: 'Violet',
	teal: 'Teal',
	emerald: 'Emerald',
	amber: 'Amber',
	crimson: 'Crimson'
};

// Approximate light-mode swatch colour for each palette. Hard-coded because
// tokens.css doesn't expose a per-palette CSS variable we can read from JS
// without temporarily mounting an off-screen element per palette.
export const PALETTE_SWATCHES: Record<Palette, string> = {
	indigo: '#4f46e5',
	violet: '#7c3aed',
	teal: '#0d9488',
	emerald: '#059669',
	amber: '#d97706',
	crimson: '#e11d48'
};

export type Theme = 'light' | 'dark';
const THEMES: readonly Theme[] = ['light', 'dark'] as const;

function isPalette(v: string | null): v is Palette {
	return v != null && (PALETTES as readonly string[]).includes(v);
}
function isTheme(v: string | null): v is Theme {
	return v != null && (THEMES as readonly string[]).includes(v);
}

class Appearance {
	palette = $state<Palette>('indigo');
	theme = $state<Theme>('light');

	// init() is safe to call multiple times — only the first call reads from
	// localStorage; subsequent ones just re-apply the current state.
	init(): void {
		if (typeof window === 'undefined') return;
		const storedPalette = window.localStorage.getItem(PALETTE_KEY);
		if (isPalette(storedPalette)) this.palette = storedPalette;
		const storedTheme = window.localStorage.getItem(THEME_KEY);
		if (isTheme(storedTheme)) this.theme = storedTheme;
		this.apply();
	}

	setPalette(p: Palette): void {
		this.palette = p;
		try {
			window.localStorage.setItem(PALETTE_KEY, p);
		} catch {
			/* private mode etc. — fine, just don't persist */
		}
		this.apply();
	}

	setTheme(t: Theme): void {
		this.theme = t;
		try {
			window.localStorage.setItem(THEME_KEY, t);
		} catch {
			/* same as above */
		}
		this.apply();
	}

	toggleTheme(): void {
		this.setTheme(this.theme === 'light' ? 'dark' : 'light');
	}

	private apply(): void {
		if (typeof document === 'undefined') return;
		document.documentElement.dataset.palette = this.palette;
		document.documentElement.dataset.theme = this.theme;
	}
}

export const appearance = new Appearance();
