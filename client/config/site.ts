export type SiteConfig = typeof siteConfig;

let navItems = [
  {
    label: "Home",
    href: "/",
  },
  {
    label: "Login",
    href: "/login",
  },
  {
    label: "Register",
    href: "/register",
  },
  {
    label: "About",
    href: "/about",
  },
]

let authNavItems = [
  {
    label: "Home",
    href: "/",
  },
  {
    label: "Send Money",
    href: "/transact",
  },
  {
    label: "Pending Games",
    href: "/games",
  },
  {
    label: "About",
    href: "/about",
  },
]

export const siteConfig = {
  name: "Trust Me Bro",
  description: "The crypto currency that relies on humanity and ethics. Just trust me bro!",
  navItems: navItems,
  authNavItems: authNavItems,
  authNavMenuItems: authNavItems,
  navMenuItems: navItems,
  links: {
    github: "https://github.com/TheScientist101/trust-me-bro",
    register: "/register",
    server: "https://trust.thescientist101.hackclub.app",
  },
};
