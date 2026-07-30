import React, { useEffect } from "react";
import { FlatList, StyleSheet, Text, TouchableOpacity, View } from "react-native";
import { useDispatch, useSelector } from "react-redux";
import Screen from "../../components/Screen";
import { useI18n } from "../../i18n/I18nContext";
import theme from "../../theme";
import { listClientsApi } from "../../redux/clients/clients.actions";
import { listProjectsApi } from "../../redux/projects/projects.actions";

const ProjectsScreen = ({ navigation }) => {
  const { t } = useI18n();
  const dispatch = useDispatch();
  const { orgId } = useSelector((state) => state.session);
  const { items: projects } = useSelector((state) => state.projects);
  const { items: clients } = useSelector((state) => state.clients);

  useEffect(() => {
    if (orgId) {
      dispatch(listClientsApi(orgId));
      dispatch(listProjectsApi(orgId));
    }
  }, [orgId, dispatch]);

  return (
    <Screen scroll={false}>
      <FlatList
        data={projects}
        keyExtractor={(p) => p.id}
        contentContainerStyle={styles.list}
        ListEmptyComponent={<Text style={styles.empty}>{t("projects.noProjects")}</Text>}
        renderItem={({ item }) => {
          const client = clients.find((c) => c.id === item.clientId);
          return (
            <TouchableOpacity style={styles.row} onPress={() => navigation.navigate("ProjectForm", { project: item })}>
              <View style={[styles.dot, { backgroundColor: item.color || theme.color.primary }]} />
              <View style={styles.rowMain}>
                <Text style={styles.rowText}>{item.name}</Text>
                {client && <Text style={styles.rowSub}>{client.name}</Text>}
              </View>
            </TouchableOpacity>
          );
        }}
      />
      <TouchableOpacity style={styles.fab} onPress={() => navigation.navigate("ProjectForm", {})}>
        <Text style={styles.fabText}>+</Text>
      </TouchableOpacity>
    </Screen>
  );
};

const styles = StyleSheet.create({
  list: {
    padding: theme.spacing(2),
    paddingBottom: theme.spacing(10),
  },
  row: {
    flexDirection: "row",
    alignItems: "center",
    paddingVertical: theme.spacing(1.5),
    borderBottomWidth: 1,
    borderBottomColor: theme.color.border,
  },
  dot: {
    width: 12,
    height: 12,
    borderRadius: 6,
    marginRight: theme.spacing(1.5),
  },
  rowMain: {
    flex: 1,
  },
  rowText: {
    fontSize: 16,
    color: theme.color.text,
  },
  rowSub: {
    fontSize: 13,
    color: theme.color.textMuted,
    marginTop: 2,
  },
  empty: {
    color: theme.color.textMuted,
    paddingVertical: theme.spacing(2),
  },
  fab: {
    position: "absolute",
    right: theme.spacing(2),
    bottom: theme.spacing(2),
    width: 56,
    height: 56,
    borderRadius: 28,
    backgroundColor: theme.color.primary,
    alignItems: "center",
    justifyContent: "center",
    elevation: 4,
  },
  fabText: {
    color: theme.color.white,
    fontSize: 28,
    lineHeight: 30,
  },
});

export default ProjectsScreen;
