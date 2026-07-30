import React, { useEffect } from "react";
import { ActivityIndicator, StyleSheet, View } from "react-native";
import { GestureHandlerRootView } from "react-native-gesture-handler";
import { SafeAreaProvider } from "react-native-safe-area-context";
import { NavigationContainer } from "@react-navigation/native";
import { createNativeStackNavigator } from "@react-navigation/native-stack";
import { createBottomTabNavigator } from "@react-navigation/bottom-tabs";
import { Provider, useDispatch, useSelector } from "react-redux";

import store from "./redux/store";
import theme from "./theme";
import { I18nProvider, useI18n } from "./i18n/I18nContext";
import { isAdminOrOwner } from "./common/permissions";
import { restoreSessionApi } from "./redux/session/session.actions";
import { listMembersApi } from "./redux/organizations/organizations.actions";

import ConnectScreen from "./screens/onboarding/ConnectScreen";
import ScanQrScreen from "./screens/onboarding/ScanQrScreen";
import ManualEntryScreen from "./screens/onboarding/ManualEntryScreen";
import OrgPickerScreen from "./screens/onboarding/OrgPickerScreen";
import TimeTrackerScreen from "./screens/time/TimeTrackerScreen";
import EditRecordScreen from "./screens/time/EditRecordScreen";
import AllDayRecordScreen from "./screens/time/AllDayRecordScreen";
import ReportsScreen from "./screens/reports/ReportsScreen";
import InvoicesScreen from "./screens/invoices/InvoicesScreen";
import PdfViewerScreen from "./screens/pdf/PdfViewerScreen";
import SettingsScreen from "./screens/settings/SettingsScreen";

const RootStack = createNativeStackNavigator();
const Tab = createBottomTabNavigator();

const LoadingScreen = () => (
  <View style={styles.loading}>
    <ActivityIndicator size="large" color={theme.color.primary} />
  </View>
);

// OnboardingNavigator: session.status === "missing" (see
// redux/session/session.reducer.js) - no stored/valid API key yet.
const OnboardingNavigator = () => {
  const { t } = useI18n();
  return (
    <RootStack.Navigator>
      <RootStack.Screen name="Connect" component={ConnectScreen} options={{ headerShown: false }} />
      <RootStack.Screen name="ScanQr" component={ScanQrScreen} options={{ title: t("onboarding.scanQr") }} />
      <RootStack.Screen
        name="ManualEntry"
        component={ManualEntryScreen}
        options={{ title: t("onboarding.enterManually") }}
      />
    </RootStack.Navigator>
  );
};

// NeedsOrgNavigator: session.status === "needsOrg" - a valid API key with no
// org_id resolved yet (manual entry without one, or an older QR).
const NeedsOrgNavigator = () => (
  <RootStack.Navigator>
    <RootStack.Screen name="PickOrganization" component={OrgPickerScreen} options={{ headerShown: false }} />
  </RootStack.Navigator>
);

// MainTabs is the connected app's tab bar. Invoices is only shown to an
// admin/owner of the current org (mirrors cwclock-ui's Slidebar.jsx
// showInvoices gate) - members/reader roles get Time tracker/Reports/
// Settings only.
const MainTabs = () => {
  const { t } = useI18n();
  const dispatch = useDispatch();
  const { orgId, user } = useSelector((state) => state.session);
  const { members } = useSelector((state) => state.organizations);

  useEffect(() => {
    if (orgId) dispatch(listMembersApi(orgId));
  }, [orgId, dispatch]);

  const showInvoices = isAdminOrOwner(user, members);

  return (
    <Tab.Navigator screenOptions={{ headerShown: true }}>
      <Tab.Screen name="TimeTracker" component={TimeTrackerScreen} options={{ title: t("timeTracker.title") }} />
      <Tab.Screen name="Reports" component={ReportsScreen} options={{ title: t("reports.title") }} />
      {showInvoices && (
        <Tab.Screen name="Invoices" component={InvoicesScreen} options={{ title: t("invoices.title") }} />
      )}
      <Tab.Screen name="Settings" component={SettingsScreen} options={{ title: t("settings.title") }} />
    </Tab.Navigator>
  );
};

// ConnectedNavigator: session.status === "connected" - the tab bar plus the
// screens pushed on top of it (record editing, PDF viewing, switching org).
const ConnectedNavigator = () => {
  const { t } = useI18n();
  return (
    <RootStack.Navigator>
      <RootStack.Screen name="Main" component={MainTabs} options={{ headerShown: false }} />
      <RootStack.Screen
        name="EditRecord"
        component={EditRecordScreen}
        options={{ title: t("timeTracker.editRecord") }}
      />
      <RootStack.Screen
        name="AllDayRecord"
        component={AllDayRecordScreen}
        options={{ title: t("timeTracker.addAllDayRecord") }}
      />
      <RootStack.Screen
        name="PdfViewer"
        component={PdfViewerScreen}
        options={({ route }) => ({ title: route.params?.title || t("pdf.title") })}
      />
      <RootStack.Screen
        name="SwitchOrganization"
        component={OrgPickerScreen}
        options={{ title: t("settings.switchOrganization") }}
      />
    </RootStack.Navigator>
  );
};

// RootNavigator switches its whole tree based on session.status, restored
// once at boot (restoreSessionApi) - no screen in any of the four
// sub-navigators needs to know about the others or navigate between them.
const RootNavigator = () => {
  const dispatch = useDispatch();
  const status = useSelector((state) => state.session.status);

  useEffect(() => {
    dispatch(restoreSessionApi());
  }, [dispatch]);

  return (
    <NavigationContainer>
      {status === "missing" && <OnboardingNavigator />}
      {status === "needsOrg" && <NeedsOrgNavigator />}
      {status === "connected" && <ConnectedNavigator />}
      {status === "restoring" && <LoadingScreen />}
    </NavigationContainer>
  );
};

const App = () => (
  <GestureHandlerRootView style={styles.flex}>
    <SafeAreaProvider>
      <Provider store={store}>
        <I18nProvider>
          <RootNavigator />
        </I18nProvider>
      </Provider>
    </SafeAreaProvider>
  </GestureHandlerRootView>
);

const styles = StyleSheet.create({
  flex: { flex: 1 },
  loading: {
    flex: 1,
    alignItems: "center",
    justifyContent: "center",
    backgroundColor: theme.color.background,
  },
});

export default App;
