export interface BrandColors {
  primary: string;
  secondary: string;
  accent: string;
}

export interface Tenant {
  id: string;
  name: string;
  slug: string;
  logo_url?: string;
  brand_colors: BrandColors;
  created_at: string;
  updated_at: string;
}

export interface UpdateBrandingPayload {
  name: string;
  logo_url?: string | null;
  brand_colors: BrandColors;
}

