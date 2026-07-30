package com.cwclock.mobile;

import com.facebook.react.ReactActivity;
import com.facebook.react.ReactActivityDelegate;
import com.facebook.react.defaults.DefaultNewArchitectureEntryPoint;
import com.facebook.react.defaults.DefaultReactActivityDelegate;

public class MainActivity extends ReactActivity {

  /**
   * Returns the name of the main component registered from JavaScript. This
   * is used to schedule rendering of the component. Must match
   * AppRegistry.registerComponent(...) in index.js.
   */
  @Override
  protected String getMainComponentName() {
    return "CWClockMobile";
  }

  /**
   * Returns the instance of the {@link ReactActivityDelegate}. We use
   * {@link DefaultReactActivityDelegate} which allows you to enable New
   * Architecture with a single boolean flag {@link #isNewArchEnabled()}.
   */
  @Override
  protected ReactActivityDelegate createReactActivityDelegate() {
    return new DefaultReactActivityDelegate(
        this, getMainComponentName(), DefaultNewArchitectureEntryPoint.getFabricEnabled());
  }
}
