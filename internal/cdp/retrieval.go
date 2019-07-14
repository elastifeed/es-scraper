package cdp

import (
	"k8s.io/klog"

	"github.com/elastifeed/es-scraper/internal/cdptab"
)

// Simple task type to save an a waiting task.
type task struct {
	Action   string
	URL      string
	Callback chan cdptab.ChromeTabReturns
}

var queue chan task

// Enqueue adds a task to the task queue which will then be run when a worker is free and a tab is ready.
func Enqueue(action, url string, callback chan cdptab.ChromeTabReturns) {
	queue <- task{
		Action:   action,
		URL:      url,
		Callback: callback,
	}
}

// The queue worker recieves tasks from the task queue and and proccesses them
func queueWorker(queue <-chan task) {
	for task := range queue {
		processRequest(task)
	}
}

// processRequest procceses a request from the task queue and runs the callback.
func processRequest(task task) {
	if tab := getFreeTab(); tab == nil {
		klog.Exitf("Nil tab assigned to worker for %s - %s", task.URL, task.Action)
		panic(tab)
	} else {
		klog.Infof("[++] Processing %s - %s on tab %d", task.URL, task.Action, tab.ID)
		defer klog.Infof("[++] Finished processing %s - %s on tab %d", task.URL, task.Action, tab.ID)
		defer tab.Ready() // When we are done, make the tab available again.

		var act func(chan cdptab.ChromeTabReturns)
		switch task.Action {
		case "all":
			act = tab.Scrape
		case "content":
			act = tab.Content
		case "screenshot":
			act = tab.Screenshot
		case "pdf":
			act = tab.Pdf
		}

		naverr := tab.Navigate(task.URL) // Navigate
		if naverr != nil {
			task.Callback <- cdptab.ChromeTabReturns{Data: nil, Err: naverr}
			return
		}
		act(task.Callback) // Execute the function.
		close(task.Callback)
	}
}

// Returns the next free tab or nil if all tabs are busy.
func getFreeTab() *cdptab.ChromeTab {
	BrowserTabs.mux.Lock()         // Lock the list to prevent to tasks being assinged to the same tab.
	defer BrowserTabs.mux.Unlock() // Unlock it when we are done

	// Find the first free tab, otherwise return nil to indicate that the request has to be queued.
	for i := range BrowserTabs.tabs {
		if BrowserTabs.tabs[i].State == cdptab.Accepting {
			klog.Infof("[+] Found ready tab: %d, -- State: %d", BrowserTabs.tabs[i].ID, BrowserTabs.tabs[i].State)
			BrowserTabs.tabs[i].Busy() // Set the found tab to busy.
			return &BrowserTabs.tabs[i]
		}
	}

	return nil
}
