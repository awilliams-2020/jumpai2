package services

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/chromedp/chromedp"
)

type LinkedInService struct {
	ctx context.Context
}

type LinkedInProfile struct {
	Name        string
	Title       string
	Company     string
	Location    string
	Description string
	Experience  []Experience
	Education   []Education
}

type Experience struct {
	Title    string
	Company  string
	Duration string
}

type Education struct {
	School    string
	Degree    string
	Field     string
	Duration  string
}

func NewLinkedInService() *LinkedInService {
	// Create a new Chrome instance
	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("headless", true),
		chromedp.Flag("disable-gpu", true),
		chromedp.Flag("no-sandbox", true),
		chromedp.Flag("disable-dev-shm-usage", true),
	)

	allocCtx, cancel := chromedp.NewExecAllocator(context.Background(), opts...)
	defer cancel()

	ctx, cancel := chromedp.NewContext(allocCtx)
	
	return &LinkedInService{
		ctx: ctx,
	}
}

func (s *LinkedInService) ScrapeProfile(ctx context.Context, url string) (*LinkedInProfile, error) {
	// Check if Chromium is available
	if _, err := exec.LookPath("chromium-browser"); err != nil {
		// Chromium is not available, return a basic profile with just the URL
		return &LinkedInProfile{
			Name:        "LinkedIn Profile",
			Title:       "Profile URL: " + url,
			Description: "Chromium is not available for scraping. Please ensure Chromium is installed.",
		}, nil
	}

	var profile LinkedInProfile

	// Extract the profile ID from the URL
	profileID := extractProfileID(url)
	if profileID == "" {
		return nil, fmt.Errorf("invalid LinkedIn URL")
	}

	// Set up Chrome options with Chromium path
	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("headless", true),
		chromedp.Flag("disable-gpu", true),
		chromedp.Flag("no-sandbox", true),
		chromedp.Flag("disable-dev-shm-usage", true),
		chromedp.ExecPath("/usr/bin/chromium-browser"),
		// Add additional options to improve stability
		chromedp.Flag("disable-web-security", true),
		chromedp.Flag("disable-features", "IsolateOrigins,site-per-process"),
		chromedp.Flag("disable-site-isolation-trials", true),
		// Add user agent to appear more like a regular browser
		chromedp.UserAgent("Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36"),
	)

	allocCtx, cancel := chromedp.NewExecAllocator(ctx, opts...)
	defer cancel()

	browserCtx, cancel := chromedp.NewContext(allocCtx)
	defer cancel()

	// Add timeout for navigation
	timeoutCtx, cancel := context.WithTimeout(browserCtx, 30*time.Second)
	defer cancel()

	// Navigate to the profile page with retries
	maxRetries := 3
	var err error
	for attempt := 1; attempt <= maxRetries; attempt++ {
		fmt.Printf("Attempt %d of %d to scrape LinkedIn profile for URL: %s\n", attempt, maxRetries, url)
		
		err = chromedp.Run(timeoutCtx,
			chromedp.Navigate(url),
			// Try multiple selectors for the profile card
			chromedp.WaitVisible(`.pv-top-card, .profile-info, .profile-header`, chromedp.ByQuery),
			chromedp.Sleep(2*time.Second), // Add a small delay to ensure page is fully loaded
			
			// Extract basic information with fallback selectors
			chromedp.Text(`.pv-top-card h1, .profile-info h1, .profile-header h1`, &profile.Name, chromedp.ByQuery),
			chromedp.Text(`.pv-top-card .text-body-medium, .profile-info .headline, .profile-header .headline`, &profile.Title, chromedp.ByQuery),
			chromedp.Text(`.pv-top-card .text-body-small, .profile-info .location, .profile-header .location`, &profile.Location, chromedp.ByQuery),
			
			// Extract about section with fallback selectors
			chromedp.Text(`#about + .pv-shared-text-with-see-more .inline-show-more-text, .about-section .description, .profile-about .content`, &profile.Description, chromedp.ByQuery),
			
			// Extract experience with more robust selectors
			chromedp.Evaluate(`
				Array.from(document.querySelectorAll('.experience-section .pv-entity__position-group, .experience-section .position, .experience .item')).map(exp => ({
					title: exp.querySelector('.pv-entity__name, .title, .position-title')?.innerText || '',
					company: exp.querySelector('.pv-entity__secondary-title, .company, .position-company')?.innerText || '',
					duration: exp.querySelector('.pv-entity__date-range, .date-range, .position-duration')?.innerText || ''
				}))
			`, &profile.Experience),
			
			// Extract education with more robust selectors
			chromedp.Evaluate(`
				Array.from(document.querySelectorAll('.education-section .pv-education-entity, .education-section .school, .education .item')).map(edu => ({
					school: edu.querySelector('.pv-entity__school-name, .school-name, .education-school')?.innerText || '',
					degree: edu.querySelector('.pv-entity__degree-name, .degree, .education-degree')?.innerText || '',
					field: edu.querySelector('.pv-entity__fos, .field, .education-field')?.innerText || '',
					duration: edu.querySelector('.pv-entity__dates, .date-range, .education-duration')?.innerText || ''
				}))
			`, &profile.Education),
		)

		if err == nil {
			fmt.Printf("Successfully scraped LinkedIn profile for: %s\n", profile.Name)
			return &profile, nil
		}

		if err == context.DeadlineExceeded {
			fmt.Printf("Attempt %d timed out after 30 seconds\n", attempt)
			if attempt < maxRetries {
				// Wait before retrying with exponential backoff
				backoff := time.Duration(attempt) * 5 * time.Second
				fmt.Printf("Waiting %v before retry...\n", backoff)
				time.Sleep(backoff)
				continue
			}
		}

		// If we get here, either we've exhausted retries or encountered a non-timeout error
		break
	}

	if err == context.DeadlineExceeded {
		return nil, fmt.Errorf("navigation timeout after 30 seconds (after %d attempts) for URL %s: %v", maxRetries, url, err)
	}
	return nil, fmt.Errorf("failed to scrape LinkedIn profile for URL %s: %v", url, err)
}

func extractProfileID(url string) string {
	// Remove any trailing slashes
	url = strings.TrimRight(url, "/")
	
	// Split by "/" and get the last part
	parts := strings.Split(url, "/")
	if len(parts) == 0 {
		return ""
	}
	
	// The profile ID is the last part of the URL
	return parts[len(parts)-1]
} 