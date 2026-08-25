# Requirements Document

## Introduction

The Desa Admin Dashboard is a web-based administrative interface for the Village (Desa) Information Management System. The dashboard provides authenticated administrators with tools to manage village information, users, content, and organizational data through a responsive Material Design interface built with Vue.js, Vite, and Tailwind CSS.

## Glossary

- **Dashboard**: The web application administrative interface
- **API**: The Desa Village Information Management System backend service
- **User**: An authenticated administrator using the Dashboard
- **Access_Token**: A JWT token used to authenticate API requests
- **Refresh_Token**: A JWT token used to obtain new Access_Tokens
- **Session**: The period during which a User maintains valid authentication
- **RBAC**: Role-Based Access Control system managing permissions
- **Banner**: A promotional image displayed on the public website
- **Berita**: News article content with text and images
- **UMKM**: Small business (Usaha Mikro Kecil Menengah) listing
- **Fasilitas**: Village facility with geographic coordinates
- **PPID**: Public Information Disclosure (Pejabat Pengelola Informasi dan Dokumentasi) document
- **Struktur**: Organizational structure information
- **Desa_Profile**: Village profile and general information
- **Form_Validation**: Client-side verification of user input before submission
- **HTTP_Error**: Non-2xx response from the API
- **Image_Upload**: Process of sending image files to the API
- **File_Upload**: Process of sending document files to the API

## Requirements

### Requirement 1: User Authentication

**User Story:** As an administrator, I want to log in to the Dashboard, so that I can access administrative functions securely.

#### Acceptance Criteria

1. THE Dashboard SHALL provide a login form accepting username and password
2. WHEN valid credentials are submitted, THE Dashboard SHALL send an authentication request to the API
3. WHEN the API returns Access_Token and Refresh_Token, THE Dashboard SHALL store both tokens securely
4. WHEN the API returns an authentication error, THE Dashboard SHALL display the error message to the User
5. WHEN authentication succeeds, THE Dashboard SHALL redirect the User to the dashboard overview page
6. THE Dashboard SHALL validate that username and password fields are not empty before submission

### Requirement 2: Session Management

**User Story:** As an administrator, I want my session to remain active, so that I don't have to log in repeatedly during normal use.

#### Acceptance Criteria

1. WHEN an Access_Token expires, THE Dashboard SHALL automatically use the Refresh_Token to obtain a new Access_Token
2. WHEN the Refresh_Token is invalid or expired, THE Dashboard SHALL redirect the User to the login page
3. WHEN the User clicks logout, THE Dashboard SHALL clear all stored tokens and redirect to the login page
4. THE Dashboard SHALL include the Access_Token in the Authorization header for all API requests
5. WHEN an API request returns a 401 status code, THE Dashboard SHALL attempt token refresh before retrying the request

### Requirement 3: Dashboard Overview

**User Story:** As an administrator, I want to see key statistics on the dashboard, so that I can quickly understand the system status.

#### Acceptance Criteria

1. WHEN the User navigates to the dashboard overview, THE Dashboard SHALL display the total count of users
2. WHEN the User navigates to the dashboard overview, THE Dashboard SHALL display the total count of published Berita articles
3. WHEN the User navigates to the dashboard overview, THE Dashboard SHALL display the total count of active UMKM listings
4. WHEN the User navigates to the dashboard overview, THE Dashboard SHALL display the total count of Fasilitas entries
5. THE Dashboard SHALL fetch statistics from the API using authenticated requests
6. WHEN the API returns an HTTP_Error, THE Dashboard SHALL display an error message to the User

### Requirement 4: User Management Interface

**User Story:** As an administrator, I want to manage user accounts, so that I can control who has access to the system.

#### Acceptance Criteria

1. THE Dashboard SHALL display a paginated list of users with username, email, role, and status
2. WHEN the User clicks create user, THE Dashboard SHALL display a form accepting username, email, password, and role
3. WHEN the User submits a valid create user form, THE Dashboard SHALL send a POST request to the API users endpoint
4. WHEN the User clicks edit on a user row, THE Dashboard SHALL display a form pre-filled with that user's current data
5. WHEN the User submits a valid edit user form, THE Dashboard SHALL send a PUT request to the API users endpoint
6. WHEN the User clicks delete on a user row, THE Dashboard SHALL display a confirmation dialog
7. WHEN the User confirms deletion, THE Dashboard SHALL send a DELETE request to the API users endpoint
8. WHEN an API operation succeeds, THE Dashboard SHALL refresh the user list and display a success message
9. THE Dashboard SHALL validate email format, password strength, and required fields before submission

### Requirement 5: Role and Permission Management

**User Story:** As an administrator, I want to manage roles and permissions, so that I can implement proper access control.

#### Acceptance Criteria

1. THE Dashboard SHALL display a list of all available roles with their assigned permissions
2. WHEN the User clicks create role, THE Dashboard SHALL display a form accepting role name and permission checkboxes
3. WHEN the User submits a valid create role form, THE Dashboard SHALL send a POST request to the API roles endpoint
4. WHEN the User clicks edit on a role, THE Dashboard SHALL display a form with current role name and selected permissions
5. WHEN the User submits a valid edit role form, THE Dashboard SHALL send a PUT request to the API roles endpoint
6. WHEN the User clicks delete on a role, THE Dashboard SHALL display a confirmation dialog
7. WHEN the User confirms role deletion, THE Dashboard SHALL send a DELETE request to the API roles endpoint
8. THE Dashboard SHALL fetch the list of available permissions from the API
9. THE Dashboard SHALL validate that role name is not empty before submission

### Requirement 6: Banner Management

**User Story:** As an administrator, I want to manage promotional banners, so that I can control what images appear on the public website.

#### Acceptance Criteria

1. THE Dashboard SHALL display a list of banners with thumbnail preview, title, display order, and status
2. WHEN the User clicks create banner, THE Dashboard SHALL display a form accepting title, image file, display order, and active status
3. WHEN the User selects an image file, THE Dashboard SHALL validate that the file is a supported image format
4. WHEN the User submits a valid create banner form, THE Dashboard SHALL perform an Image_Upload to the API banners endpoint
5. WHEN the User clicks edit on a banner, THE Dashboard SHALL display a form with current banner data and image preview
6. WHEN the User submits a valid edit banner form, THE Dashboard SHALL send a PUT request to the API banners endpoint
7. WHEN the User clicks delete on a banner, THE Dashboard SHALL display a confirmation dialog
8. WHEN the User confirms banner deletion, THE Dashboard SHALL send a DELETE request to the API banners endpoint
9. THE Dashboard SHALL support image formats including JPEG, PNG, and WebP
10. THE Dashboard SHALL display image file size and validate maximum file size before upload

### Requirement 7: News Article Management

**User Story:** As an administrator, I want to manage news articles, so that I can publish and update village news content.

#### Acceptance Criteria

1. THE Dashboard SHALL display a paginated list of Berita articles with title, author, publication date, and status
2. WHEN the User clicks create article, THE Dashboard SHALL display a form accepting title, content, featured image, category, and publication status
3. WHEN the User enters article content, THE Dashboard SHALL provide a rich text editor with formatting options
4. WHEN the User selects a featured image, THE Dashboard SHALL validate that the file is a supported image format
5. WHEN the User submits a valid create article form, THE Dashboard SHALL perform an Image_Upload to the API berita endpoint
6. WHEN the User clicks edit on an article, THE Dashboard SHALL display a form pre-filled with current article data
7. WHEN the User submits a valid edit article form, THE Dashboard SHALL send a PUT request to the API berita endpoint
8. WHEN the User clicks delete on an article, THE Dashboard SHALL display a confirmation dialog
9. WHEN the User confirms article deletion, THE Dashboard SHALL send a DELETE request to the API berita endpoint
10. THE Dashboard SHALL validate that title and content fields are not empty before submission
11. THE Dashboard SHALL support draft and published status for articles

### Requirement 8: UMKM Management

**User Story:** As an administrator, I want to manage small business listings, so that I can promote local enterprises.

#### Acceptance Criteria

1. THE Dashboard SHALL display a paginated list of UMKM entries with business name, category, owner, and contact information
2. WHEN the User clicks create UMKM, THE Dashboard SHALL display a form accepting business name, description, category, owner name, phone number, address, and optional image
3. WHEN the User selects a business image, THE Dashboard SHALL validate that the file is a supported image format
4. WHEN the User submits a valid create UMKM form, THE Dashboard SHALL send a POST request to the API umkm endpoint
5. WHEN the User clicks edit on a UMKM entry, THE Dashboard SHALL display a form pre-filled with current business data
6. WHEN the User submits a valid edit UMKM form, THE Dashboard SHALL send a PUT request to the API umkm endpoint
7. WHEN the User clicks delete on a UMKM entry, THE Dashboard SHALL display a confirmation dialog
8. WHEN the User confirms UMKM deletion, THE Dashboard SHALL send a DELETE request to the API umkm endpoint
9. THE Dashboard SHALL validate phone number format and required fields before submission

### Requirement 9: Facilities Management

**User Story:** As an administrator, I want to manage village facilities with location data, so that residents can find public services.

#### Acceptance Criteria

1. THE Dashboard SHALL display a list of Fasilitas entries with name, type, address, and coordinates
2. WHEN the User clicks create facility, THE Dashboard SHALL display a form accepting name, type, description, address, latitude, and longitude
3. WHEN the User enters coordinates, THE Dashboard SHALL display a map preview showing the facility location
4. WHEN the User clicks on the map, THE Dashboard SHALL update the latitude and longitude fields with the clicked coordinates
5. WHEN the User submits a valid create facility form, THE Dashboard SHALL send a POST request to the API fasilitas endpoint
6. WHEN the User clicks edit on a facility, THE Dashboard SHALL display a form pre-filled with current facility data and map location
7. WHEN the User submits a valid edit facility form, THE Dashboard SHALL send a PUT request to the API fasilitas endpoint
8. WHEN the User clicks delete on a facility, THE Dashboard SHALL display a confirmation dialog
9. WHEN the User confirms facility deletion, THE Dashboard SHALL send a DELETE request to the API fasilitas endpoint
10. THE Dashboard SHALL validate that latitude is between -90 and 90 degrees
11. THE Dashboard SHALL validate that longitude is between -180 and 180 degrees
12. THE Dashboard SHALL integrate a map library for coordinate selection and visualization

### Requirement 10: PPID Document Management

**User Story:** As an administrator, I want to manage public information disclosure documents, so that I can maintain transparency requirements.

#### Acceptance Criteria

1. THE Dashboard SHALL display a list of PPID documents with title, category, upload date, and file size
2. WHEN the User clicks create document, THE Dashboard SHALL display a form accepting title, description, category, and file upload
3. WHEN the User selects a document file, THE Dashboard SHALL validate that the file is a supported document format
4. WHEN the User submits a valid create document form, THE Dashboard SHALL perform a File_Upload to the API ppid endpoint
5. WHEN the User clicks edit on a document, THE Dashboard SHALL display a form with current document metadata
6. WHEN the User submits a valid edit document form, THE Dashboard SHALL send a PUT request to the API ppid endpoint
7. WHEN the User clicks delete on a document, THE Dashboard SHALL display a confirmation dialog
8. WHEN the User confirms document deletion, THE Dashboard SHALL send a DELETE request to the API ppid endpoint
9. THE Dashboard SHALL support document formats including PDF, DOC, DOCX, and XLS
10. THE Dashboard SHALL display file size and validate maximum file size before upload

### Requirement 11: Organizational Structure Management

**User Story:** As an administrator, I want to manage the organizational structure, so that residents can see village leadership and staff.

#### Acceptance Criteria

1. THE Dashboard SHALL display the current Struktur organizational hierarchy
2. WHEN the User clicks create position, THE Dashboard SHALL display a form accepting position title, person name, photo, and parent position
3. WHEN the User selects a photo, THE Dashboard SHALL validate that the file is a supported image format
4. WHEN the User submits a valid create position form, THE Dashboard SHALL send a POST request to the API struktur endpoint
5. WHEN the User clicks edit on a position, THE Dashboard SHALL display a form pre-filled with current position data
6. WHEN the User submits a valid edit position form, THE Dashboard SHALL send a PUT request to the API struktur endpoint
7. WHEN the User clicks delete on a position, THE Dashboard SHALL display a confirmation dialog
8. WHEN the User confirms position deletion, THE Dashboard SHALL send a DELETE request to the API struktur endpoint
9. THE Dashboard SHALL display the organizational structure as a hierarchical tree or chart
10. THE Dashboard SHALL validate that position title and person name are not empty before submission

### Requirement 12: Village Profile Management

**User Story:** As an administrator, I want to manage the village profile information, so that I can keep general village data up to date.

#### Acceptance Criteria

1. THE Dashboard SHALL display the current Desa_Profile information including village name, address, population, and description
2. WHEN the User clicks edit profile, THE Dashboard SHALL display a form pre-filled with current profile data
3. WHEN the User submits a valid edit profile form, THE Dashboard SHALL send a PUT request to the API desa endpoint
4. WHEN the profile update succeeds, THE Dashboard SHALL display a success message and refresh the profile display
5. THE Dashboard SHALL validate that village name and address fields are not empty before submission
6. THE Dashboard SHALL validate that population is a positive integer

### Requirement 13: Responsive Design

**User Story:** As an administrator, I want to use the Dashboard on different devices, so that I can manage the system from desktop or mobile.

#### Acceptance Criteria

1. THE Dashboard SHALL render correctly on desktop screens with width greater than 1024 pixels
2. THE Dashboard SHALL render correctly on tablet screens with width between 768 and 1024 pixels
3. THE Dashboard SHALL render correctly on mobile screens with width less than 768 pixels
4. WHEN the screen width is less than 768 pixels, THE Dashboard SHALL display a collapsible navigation menu
5. THE Dashboard SHALL use Tailwind CSS responsive utility classes for layout adaptation
6. THE Dashboard SHALL implement Material Design principles for visual consistency
7. THE Dashboard SHALL ensure all interactive elements have touch-friendly sizes on mobile devices

### Requirement 14: Form Validation and Error Handling

**User Story:** As an administrator, I want clear feedback on form errors, so that I can correct mistakes quickly.

#### Acceptance Criteria

1. WHEN a User submits a form with empty required fields, THE Dashboard SHALL display field-specific error messages
2. WHEN a User enters invalid data format, THE Dashboard SHALL display format-specific error messages
3. WHEN the API returns a validation error, THE Dashboard SHALL display the API error message near the relevant form field
4. WHEN the API returns a server error, THE Dashboard SHALL display a general error notification
5. THE Dashboard SHALL disable the submit button while a form submission is in progress
6. WHEN a form submission completes, THE Dashboard SHALL re-enable the submit button
7. THE Dashboard SHALL display a loading indicator during API requests
8. THE Dashboard SHALL implement Form_Validation for all input fields before API submission

### Requirement 15: Navigation and Routing

**User Story:** As an administrator, I want intuitive navigation between sections, so that I can access different management functions easily.

#### Acceptance Criteria

1. THE Dashboard SHALL provide a sidebar navigation menu with links to all major sections
2. THE Dashboard SHALL highlight the current active section in the navigation menu
3. WHEN the User clicks a navigation link, THE Dashboard SHALL navigate to the corresponding route without page reload
4. THE Dashboard SHALL use Vue Router for client-side routing
5. WHEN the User is not authenticated, THE Dashboard SHALL redirect protected routes to the login page
6. THE Dashboard SHALL display breadcrumb navigation showing the current location hierarchy
7. THE Dashboard SHALL maintain browser history for back and forward navigation

### Requirement 16: Data Pagination

**User Story:** As an administrator, I want paginated lists for large datasets, so that pages load quickly and data is manageable.

#### Acceptance Criteria

1. WHEN a list contains more than 20 items, THE Dashboard SHALL display pagination controls
2. THE Dashboard SHALL display page numbers, previous button, and next button in pagination controls
3. WHEN the User clicks a page number, THE Dashboard SHALL fetch and display that page of data
4. WHEN the User clicks previous, THE Dashboard SHALL navigate to the previous page if not on the first page
5. WHEN the User clicks next, THE Dashboard SHALL navigate to the next page if not on the last page
6. THE Dashboard SHALL display the current page number and total page count
7. THE Dashboard SHALL fetch paginated data from the API using page and limit query parameters

### Requirement 17: Search and Filtering

**User Story:** As an administrator, I want to search and filter lists, so that I can find specific items quickly.

#### Acceptance Criteria

1. THE Dashboard SHALL provide a search input field for user, Berita, UMKM, and Fasilitas lists
2. WHEN the User enters text in the search field, THE Dashboard SHALL filter the displayed list to matching items
3. THE Dashboard SHALL implement search with a debounce delay of 300 milliseconds
4. WHEN the User clears the search field, THE Dashboard SHALL display the full unfiltered list
5. WHERE filtering by category is available, THE Dashboard SHALL provide a category dropdown filter
6. WHEN the User selects a category filter, THE Dashboard SHALL display only items in that category
7. THE Dashboard SHALL send search and filter parameters to the API as query parameters

### Requirement 18: Image Preview and Optimization

**User Story:** As an administrator, I want to preview images before upload, so that I can verify the correct file is selected.

#### Acceptance Criteria

1. WHEN the User selects an image file, THE Dashboard SHALL display a preview of the image
2. THE Dashboard SHALL display the image filename and file size
3. WHEN the image file size exceeds 5 megabytes, THE Dashboard SHALL display a warning message
4. WHEN the image dimensions exceed 4000 pixels in width or height, THE Dashboard SHALL display a warning message
5. THE Dashboard SHALL validate image MIME type matches supported formats
6. WHEN the User removes the selected image, THE Dashboard SHALL clear the preview and file input

### Requirement 19: Confirmation Dialogs

**User Story:** As an administrator, I want confirmation before destructive actions, so that I can prevent accidental deletions.

#### Acceptance Criteria

1. WHEN the User initiates a delete action, THE Dashboard SHALL display a modal confirmation dialog
2. THE Dashboard SHALL display the name or title of the item being deleted in the confirmation message
3. WHEN the User clicks confirm in the dialog, THE Dashboard SHALL proceed with the delete request
4. WHEN the User clicks cancel in the dialog, THE Dashboard SHALL close the dialog without performing the delete
5. THE Dashboard SHALL prevent interaction with the underlying page while the confirmation dialog is open

### Requirement 20: Success and Error Notifications

**User Story:** As an administrator, I want clear notifications of operation results, so that I know when actions succeed or fail.

#### Acceptance Criteria

1. WHEN an API operation succeeds, THE Dashboard SHALL display a success notification with a descriptive message
2. WHEN an API operation fails, THE Dashboard SHALL display an error notification with the error message
3. THE Dashboard SHALL automatically dismiss success notifications after 3 seconds
4. THE Dashboard SHALL keep error notifications visible until the User dismisses them
5. THE Dashboard SHALL display notifications in a consistent location on the screen
6. THE Dashboard SHALL allow multiple notifications to stack vertically

### Requirement 21: Password Reset Flow

**User Story:** As an administrator, I want to reset my password if forgotten, so that I can regain access to my account.

#### Acceptance Criteria

1. THE Dashboard SHALL provide a "Forgot Password" link on the login page
2. WHEN the User clicks "Forgot Password", THE Dashboard SHALL display a form requesting email address
3. WHEN the User submits a valid email, THE Dashboard SHALL send a password reset request to the API
4. WHEN the API confirms the reset email was sent, THE Dashboard SHALL display a confirmation message
5. WHEN the User follows a password reset link, THE Dashboard SHALL display a form accepting new password and confirmation
6. WHEN the User submits matching valid passwords, THE Dashboard SHALL send the password reset to the API
7. WHEN password reset succeeds, THE Dashboard SHALL redirect to the login page with a success message
8. THE Dashboard SHALL validate that new password meets minimum strength requirements
9. THE Dashboard SHALL validate that password and confirmation fields match

### Requirement 22: API Request Configuration

**User Story:** As a developer, I want centralized API configuration, so that the Dashboard can connect to different API environments.

#### Acceptance Criteria

1. THE Dashboard SHALL read the API base URL from an environment configuration file
2. THE Dashboard SHALL configure Axios with the API base URL as the default baseURL
3. THE Dashboard SHALL configure Axios to include the Access_Token in request headers automatically
4. THE Dashboard SHALL configure Axios to handle token refresh on 401 responses
5. THE Dashboard SHALL configure Axios request timeout to 30 seconds
6. THE Dashboard SHALL configure Axios to parse JSON responses automatically

### Requirement 23: Loading States

**User Story:** As an administrator, I want visual feedback during data loading, so that I know the system is working.

#### Acceptance Criteria

1. WHEN the Dashboard fetches data from the API, THE Dashboard SHALL display a loading spinner or skeleton screen
2. WHEN data loading completes, THE Dashboard SHALL hide the loading indicator and display the data
3. WHEN a form submission is in progress, THE Dashboard SHALL display a loading indicator on the submit button
4. WHEN a form submission completes, THE Dashboard SHALL hide the loading indicator
5. THE Dashboard SHALL disable interactive elements while loading is in progress

### Requirement 24: Accessibility Compliance

**User Story:** As an administrator with accessibility needs, I want the Dashboard to be usable with assistive technologies, so that I can perform my job effectively.

#### Acceptance Criteria

1. THE Dashboard SHALL provide text alternatives for all images using alt attributes
2. THE Dashboard SHALL ensure all interactive elements are keyboard accessible
3. THE Dashboard SHALL provide visible focus indicators for keyboard navigation
4. THE Dashboard SHALL use semantic HTML elements for proper structure
5. THE Dashboard SHALL ensure color contrast ratios meet WCAG AA standards
6. THE Dashboard SHALL provide ARIA labels for icon-only buttons
7. THE Dashboard SHALL announce dynamic content changes to screen readers using ARIA live regions

### Requirement 25: Build and Development Configuration

**User Story:** As a developer, I want a configured build system, so that I can develop and deploy the Dashboard efficiently.

#### Acceptance Criteria

1. THE Dashboard SHALL use Vite as the build tool and development server
2. THE Dashboard SHALL configure Vite with Vue.js plugin support
3. THE Dashboard SHALL configure Tailwind CSS for utility-first styling
4. THE Dashboard SHALL configure PostCSS for CSS processing
5. THE Dashboard SHALL provide a development script that starts a local development server with hot module replacement
6. THE Dashboard SHALL provide a build script that generates optimized production assets
7. THE Dashboard SHALL configure environment variables for API URL and other configuration
8. THE Dashboard SHALL generate source maps for debugging in development mode

### Requirement 26: Code Organization

**User Story:** As a developer, I want well-organized code structure, so that the Dashboard is maintainable and scalable.

#### Acceptance Criteria

1. THE Dashboard SHALL organize Vue components in a components directory
2. THE Dashboard SHALL organize page components in a views or pages directory
3. THE Dashboard SHALL organize API client code in a services or api directory
4. THE Dashboard SHALL organize routing configuration in a router directory
5. THE Dashboard SHALL organize state management in a store directory if using Vuex or Pinia
6. THE Dashboard SHALL organize utility functions in a utils directory
7. THE Dashboard SHALL organize reusable composables in a composables directory
8. THE Dashboard SHALL use consistent naming conventions for files and components

### Requirement 27: Git Version Control

**User Story:** As a developer, I want version control for the Dashboard code, so that I can track changes and collaborate effectively.

#### Acceptance Criteria

1. THE Dashboard SHALL initialize a Git repository in the project root
2. THE Dashboard SHALL include a .gitignore file excluding node_modules, dist, and environment files
3. THE Dashboard SHALL commit the initial project structure with a descriptive commit message
4. THE Dashboard SHALL commit each major feature implementation separately with descriptive messages
5. THE Dashboard SHALL use conventional commit message format for clarity

## Optional Requirements

### Optional Requirement 1: Multi-language Support

**User Story:** As an administrator, I want to switch the Dashboard language, so that I can use it in my preferred language.

#### Acceptance Criteria

1. WHERE multi-language support is enabled, THE Dashboard SHALL provide a language selector in the user interface
2. WHERE multi-language support is enabled, WHEN the User selects a language, THE Dashboard SHALL update all interface text to the selected language
3. WHERE multi-language support is enabled, THE Dashboard SHALL persist the language preference in browser storage
4. WHERE multi-language support is enabled, THE Dashboard SHALL use Vue I18n for internationalization
5. WHERE multi-language support is enabled, THE Dashboard SHALL support at least Indonesian and English languages

### Optional Requirement 2: Theme Switching

**User Story:** As an administrator, I want to switch between light and dark themes, so that I can use the Dashboard comfortably in different lighting conditions.

#### Acceptance Criteria

1. WHERE theme switching is enabled, THE Dashboard SHALL provide a theme toggle button in the user interface
2. WHERE theme switching is enabled, WHEN the User clicks the theme toggle, THE Dashboard SHALL switch between light and dark themes
3. WHERE theme switching is enabled, THE Dashboard SHALL persist the theme preference in browser storage
4. WHERE theme switching is enabled, THE Dashboard SHALL apply theme-appropriate colors to all components
5. WHERE theme switching is enabled, THE Dashboard SHALL use Tailwind CSS dark mode utilities for theme implementation
6. WHERE theme switching is enabled, THE Dashboard SHALL respect the user's system theme preference on first visit
