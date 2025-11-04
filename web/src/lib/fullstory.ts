import { FullStory, init }from '@fullstory/browser';

class FullStoryService {
  private initialized = false

  /**
   * Initialize FullStory with organization ID
   */
  init(orgId: string): void {
    if (this.initialized) {
      return
    }

    try {
      init({ orgId })
      this.initialized = true
      console.log('FullStory initialized')
    } catch (error) {
      console.error('Failed to initialize FullStory:', error)
    }
  }

  /**
   * Identify the current user
   */
  identify(userId: string): void {
    if (!this.initialized) {
      return
    }

    try {
      FullStory('setIdentity', userId)
    } catch (error) {
      console.error('Failed to identify user in FullStory:', error)
    }
  }
}

// Export a singleton instance
export const fullStory = new FullStoryService()